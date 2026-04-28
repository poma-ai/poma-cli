package client

import (
	"strings"
	"testing"
)

// helpers

func mkChunk(idx int, content string, depth int, fileID string) ChunkInput {
	return ChunkInput{ChunkIndex: idx, Content: content, Depth: depth, FileID: fileID}
}

func mkCodeChunk(idx int, content string, depth int, fileID, code string) ChunkInput {
	return ChunkInput{ChunkIndex: idx, Content: content, Depth: depth, FileID: fileID, Code: code}
}

func mkCS(fileID string, chunks ...int) ChunksetInput {
	return ChunksetInput{FileID: fileID, Chunks: chunks}
}

// Test 1: single doc, consecutive chunks — no gap marker
func TestGenerateCheatsheets_SingleDocConsecutive(t *testing.T) {
	chunks := []ChunkInput{
		mkChunk(0, "line0", 0, ""),
		mkChunk(1, "line1", 0, ""),
		mkChunk(2, "line2", 0, ""),
	}
	chunksets := []ChunksetInput{mkCS("", 0, 1, 2)}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	if result[0].FileID != "single_doc" {
		t.Errorf("expected file_id 'single_doc', got %q", result[0].FileID)
	}
	if strings.Contains(result[0].Content, "[…]") {
		t.Errorf("expected no gap marker, got: %q", result[0].Content)
	}
	for _, line := range []string{"line0", "line1", "line2"} {
		if !strings.Contains(result[0].Content, line) {
			t.Errorf("expected content to contain %q", line)
		}
	}
}

// Test 2: gap between chunks — gap marker present
func TestGenerateCheatsheets_Gap(t *testing.T) {
	// 5 chunks at same depth; chunkset refs [0, 4]
	// parent expansion: 0 and 4 are both roots (depth 0) so no parents
	// child expansion: no children at deeper depth since all depth=0
	// result: just chunks 0 and 4 with a gap marker
	chunks := []ChunkInput{
		mkChunk(0, "a", 0, ""),
		mkChunk(1, "b", 0, ""),
		mkChunk(2, "c", 0, ""),
		mkChunk(3, "d", 0, ""),
		mkChunk(4, "e", 0, ""),
	}
	chunksets := []ChunksetInput{mkCS("", 0, 4)}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	if !strings.Contains(result[0].Content, "[…]") {
		t.Errorf("expected gap marker '[…]', got: %q", result[0].Content)
	}
}

// Test 3: multi-doc — two file_ids, separate cheatsheets
func TestGenerateCheatsheets_MultiDoc(t *testing.T) {
	chunks := []ChunkInput{
		mkChunk(0, "doc_a_0", 0, "doc_a"),
		mkChunk(1, "doc_a_1", 0, "doc_a"),
		mkChunk(0, "doc_b_0", 0, "doc_b"),
		mkChunk(1, "doc_b_1", 0, "doc_b"),
	}
	chunksets := []ChunksetInput{
		mkCS("doc_a", 0, 1),
		mkCS("doc_b", 0, 1),
	}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 cheatsheets, got %d", len(result))
	}
	byID := make(map[string]string)
	for _, cs := range result {
		byID[cs.FileID] = cs.Content
	}
	if !strings.Contains(byID["doc_a"], "doc_a_0") {
		t.Errorf("doc_a cheatsheet missing expected content")
	}
	if !strings.Contains(byID["doc_b"], "doc_b_0") {
		t.Errorf("doc_b cheatsheet missing expected content")
	}
	if strings.Contains(byID["doc_a"], "doc_b") {
		t.Errorf("doc_a cheatsheet contains doc_b content")
	}
}

// Test 4: code indent — depth rebased to 1 level → 4-space indent
func TestGenerateCheatsheets_CodeIndent(t *testing.T) {
	chunks := []ChunkInput{
		mkCodeChunk(0, "code_line", 2, "", "block1"),
		mkCodeChunk(1, "code_line2", 3, "", "block1"),
	}
	chunksets := []ChunksetInput{mkCS("", 0, 1)}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	// block min depth = 2; chunk 0 rebased = 0 (no indent); chunk 1 rebased = 1 (4 spaces)
	if strings.Contains(result[0].Content, "    code_line\n") {
		t.Errorf("chunk at min depth should not be indented, got: %q", result[0].Content)
	}
	if !strings.Contains(result[0].Content, "    code_line2") {
		t.Errorf("chunk at depth+1 should be indented by 4 spaces, got: %q", result[0].Content)
	}
}

// Test 5: duplicate chunk_index within same file_id → error
func TestGenerateCheatsheets_DuplicateChunkIndex(t *testing.T) {
	chunks := []ChunkInput{
		mkChunk(0, "a", 0, ""),
		mkChunk(0, "b", 0, ""), // duplicate
	}
	chunksets := []ChunksetInput{mkCS("", 0)}

	_, err := GenerateCheatsheets(chunksets, chunks)
	if err == nil {
		t.Fatal("expected error for duplicate chunk_index, got nil")
	}
}

// Test 6: file_id in chunksets not in all_chunks → error
func TestGenerateCheatsheets_MissingFileID(t *testing.T) {
	chunks := []ChunkInput{
		mkChunk(0, "a", 0, "doc_a"),
	}
	chunksets := []ChunksetInput{mkCS("doc_b", 0)} // doc_b not in chunks

	_, err := GenerateCheatsheets(chunksets, chunks)
	if err == nil {
		t.Fatal("expected error for missing file_id, got nil")
	}
}

// Test 7: single_doc fallback — chunks with no file_id
func TestGenerateCheatsheets_SingleDocFallback(t *testing.T) {
	chunks := []ChunkInput{
		{ChunkIndex: 0, Content: "hello", Depth: 0}, // no FileID
	}
	chunksets := []ChunksetInput{{Chunks: []int{0}}} // no FileID

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	if result[0].FileID != "single_doc" {
		t.Errorf("expected file_id 'single_doc', got %q", result[0].FileID)
	}
	if result[0].Content != "hello" {
		t.Errorf("expected content 'hello', got %q", result[0].Content)
	}
}

// Test 8: parent/child expansion — chunkset refs a mid-level node; ancestors and subtree included.
// Siblings at the same depth are NOT included (child expansion requires strictly deeper depth).
func TestGenerateCheatsheets_ParentChildExpansion(t *testing.T) {
	// Tree: 0(depth0) -> 1(depth1) -> 2(depth2) -> 3(depth3)
	//                                            -> 4(depth3)
	// Chunkset refs [2]; expect 0,1,2,3,4 (parents 0,1 + hit 2 + children 3,4)
	chunks := []ChunkInput{
		mkChunk(0, "root", 0, ""),
		mkChunk(1, "section", 1, ""),
		mkChunk(2, "hit", 2, ""),
		mkChunk(3, "child_a", 3, ""),
		mkChunk(4, "child_b", 3, ""),
	}
	chunksets := []ChunksetInput{mkCS("", 2)}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	for _, want := range []string{"root", "section", "hit", "child_a", "child_b"} {
		if !strings.Contains(result[0].Content, want) {
			t.Errorf("expected %q in cheatsheet content, got: %q", want, result[0].Content)
		}
	}
}

// Test 9: chunkIDs referencing indices absent from all_chunks — silently dropped
func TestGenerateCheatsheets_UnknownChunkIDSilentlyDropped(t *testing.T) {
	chunks := []ChunkInput{
		mkChunk(0, "present", 0, ""),
	}
	// Chunkset references IDs 0 (exists) and 99 (does not exist)
	chunksets := []ChunksetInput{mkCS("", 0, 99)}

	result, err := GenerateCheatsheets(chunksets, chunks)
	if err != nil {
		t.Fatalf("expected no error for unknown chunk ID, got: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 cheatsheet, got %d", len(result))
	}
	if !strings.Contains(result[0].Content, "present") {
		t.Errorf("expected 'present' in cheatsheet content, got: %q", result[0].Content)
	}
	// ID 99 is silently ignored — no "[99]" or error text in output
	if strings.Contains(result[0].Content, "99") {
		t.Errorf("unexpected content from unknown chunk ID in output: %q", result[0].Content)
	}
}
