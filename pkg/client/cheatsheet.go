package client

import (
	"fmt"
	"sort"
	"strings"
)

// Cheatsheet is a generated cheatsheet for one document.
type Cheatsheet struct {
	FileID  string `json:"file_id"`
	Content string `json:"content"`
}

// ChunkInput is a single chunk as provided in the input JSON.
type ChunkInput struct {
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
	Depth      int    `json:"depth"`
	FileID     string `json:"file_id"`
	Code       string `json:"code"`
}

// ChunksetInput is a single chunkset as provided in the input JSON.
type ChunksetInput struct {
	FileID         string `json:"file_id"`
	ChunksetIndex  int    `json:"chunkset_index"`
	Chunks         []int  `json:"chunks"`
	ToEmbed        string `json:"to_embed"`
}

// CheatsheetRequest is the top-level input JSON for cheatsheet generation.
type CheatsheetRequest struct {
	RelevantChunksets []ChunksetInput `json:"relevant_chunksets"`
	AllChunks         []ChunkInput    `json:"all_chunks"`
}

// retrievalChunk is an internal representation used during assembly.
type retrievalChunk struct {
	index        int
	fileID       string
	content      string
	depthRebased *int // nil for non-code chunks
}

// GenerateCheatsheets generates one cheatsheet per document from the given
// relevant chunksets and all chunks. It is a Go port of generate_cheatsheets
// in bin/sdk/retrieval.py.
func GenerateCheatsheets(relevantChunksets []ChunksetInput, allChunks []ChunkInput) ([]Cheatsheet, error) {
	if len(relevantChunksets) == 0 {
		return nil, fmt.Errorf("relevant_chunksets must be non-empty")
	}
	if len(allChunks) == 0 {
		return nil, fmt.Errorf("all_chunks must be non-empty")
	}

	// Step 1: group chunks by file_id, default "single_doc"
	docChunks := make(map[string][]ChunkInput)
	for _, ch := range allChunks {
		fid := fileIDOrDefault(ch.FileID)
		docChunks[fid] = append(docChunks[fid], ch)
	}

	// Check for duplicate chunk_index within each file_id
	for fid, chunks := range docChunks {
		seen := make(map[int]struct{}, len(chunks))
		for _, ch := range chunks {
			if _, dup := seen[ch.ChunkIndex]; dup {
				return nil, fmt.Errorf("duplicate chunk_index %d for file_id %q", ch.ChunkIndex, fid)
			}
			seen[ch.ChunkIndex] = struct{}{}
		}
	}

	// Step 2: group relevant chunksets by file_id
	relevantPerDoc := make(map[string][]ChunksetInput)
	for _, cs := range relevantChunksets {
		if cs.Chunks == nil {
			return nil, fmt.Errorf("chunkset at index %d is missing required 'chunks' key", cs.ChunksetIndex)
		}
		fid := fileIDOrDefault(cs.FileID)
		relevantPerDoc[fid] = append(relevantPerDoc[fid], cs)
	}

	// Validate that every file_id in chunksets exists in chunks
	for fid := range relevantPerDoc {
		if _, ok := docChunks[fid]; !ok {
			return nil, fmt.Errorf("chunksets contain file_id %q which is not present in all_chunks", fid)
		}
	}

	// Step 3: collect relevant retrieval chunks per document
	var contentChunks []retrievalChunk
	for fid, chunksets := range relevantPerDoc {
		var chunkIDs []int
		for _, cs := range chunksets {
			chunkIDs = append(chunkIDs, cs.Chunks...)
		}
		relevant, err := getRelevantChunksForIDs(chunkIDs, docChunks[fid])
		if err != nil {
			return nil, err
		}
		rc := toRetrievalChunks(relevant, fid)
		contentChunks = append(contentChunks, rc...)
	}

	// Step 4: assemble cheatsheets
	return cheatsheetsFromChunks(contentChunks), nil
}

// fileIDOrDefault returns "single_doc" when fid is blank.
func fileIDOrDefault(fid string) string {
	if fid == "" {
		return "single_doc"
	}
	return fid
}

// toRetrievalChunks converts ChunkInput slices to retrievalChunk, computing
// rebased depth for code-block chunks. The fileID parameter is used as a
// fallback when a chunk's own FileID is blank.
func toRetrievalChunks(chunks []ChunkInput, fileID string) []retrievalChunk {
	// Compute minimum depth per code block id
	minDepthPerBlock := make(map[string]int)
	for _, ch := range chunks {
		if ch.Code == "" {
			continue
		}
		if cur, ok := minDepthPerBlock[ch.Code]; !ok || ch.Depth < cur {
			minDepthPerBlock[ch.Code] = ch.Depth
		}
	}

	result := make([]retrievalChunk, 0, len(chunks))
	for _, ch := range chunks {
		fid := ch.FileID
		if fid == "" {
			fid = fileID
		}
		rc := retrievalChunk{
			index:   ch.ChunkIndex,
			fileID:  fileIDOrDefault(fid),
			content: ch.Content,
		}
		if ch.Code != "" {
			if minDepth, ok := minDepthPerBlock[ch.Code]; ok {
				rebased := ch.Depth - minDepth
				if rebased < 0 {
					rebased = 0
				}
				rc.depthRebased = &rebased
			}
		}
		result = append(result, rc)
	}
	return result
}

// cheatsheetsFromChunks assembles cheatsheets from sorted retrieval chunks.
func cheatsheetsFromChunks(chunks []retrievalChunk) []Cheatsheet {
	if len(chunks) == 0 {
		return nil
	}

	// Sort by (fileID, index)
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].fileID != chunks[j].fileID {
			return chunks[i].fileID < chunks[j].fileID
		}
		return chunks[i].index < chunks[j].index
	})

	var cheatsheets []Cheatsheet
	var curFileID string
	var sb strings.Builder
	lastIndex := -1

	flush := func() {
		if curFileID != "" {
			cheatsheets = append(cheatsheets, Cheatsheet{
				FileID:  curFileID,
				Content: sb.String(),
			})
		}
	}

	for _, ch := range chunks {
		formatted := formatChunkContent(ch)

		if ch.fileID != curFileID {
			flush()
			curFileID = ch.fileID
			sb.Reset()
			sb.WriteString(formatted)
			lastIndex = ch.index
		} else {
			if ch.index == lastIndex+1 {
				sb.WriteString("\n")
			} else {
				sb.WriteString("\n[…]\n")
			}
			sb.WriteString(formatted)
			lastIndex = ch.index
		}
	}
	flush()

	return cheatsheets
}

// formatChunkContent returns the chunk's content, indented for code blocks.
func formatChunkContent(ch retrievalChunk) string {
	if ch.depthRebased == nil || *ch.depthRebased == 0 {
		return ch.content
	}
	indent := strings.Repeat("    ", *ch.depthRebased)
	return indent + ch.content
}

// getRelevantChunksForIDs returns the expanded set of relevant chunks
// (hits + subtree children + ancestor parents) in ascending chunk_index order.
// Port of _get_relevant_chunks_for_ids in bin/sdk/retrieval.py.
func getRelevantChunksForIDs(chunkIDs []int, chunks []ChunkInput) ([]ChunkInput, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	// Build sorted list and lookup maps
	sorted := make([]ChunkInput, len(chunks))
	copy(sorted, chunks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ChunkIndex < sorted[j].ChunkIndex
	})

	indexToChunk := make(map[int]ChunkInput, len(sorted))
	indexToDepth := make(map[int]int, len(sorted))
	for _, ch := range sorted {
		indexToChunk[ch.ChunkIndex] = ch
		indexToDepth[ch.ChunkIndex] = ch.Depth
	}

	candidate := make(map[int]struct{}, len(chunkIDs))
	for _, id := range chunkIDs {
		candidate[id] = struct{}{}
	}

	// Find relatively deepest: remove idx1 that is an ancestor of any idx2 in the set
	relativelyDeepest := make(map[int]struct{}, len(candidate))
	for id := range candidate {
		relativelyDeepest[id] = struct{}{}
	}
	for idx1 := range candidate {
		for idx2 := range candidate {
			if idx1 != idx2 && isAncestor(idx1, idx2, sorted, indexToDepth) {
				delete(relativelyDeepest, idx1)
				break
			}
		}
	}

	// Expand by children for each relatively deepest index
	allIndices := make(map[int]struct{}, len(chunkIDs))
	for id := range candidate {
		allIndices[id] = struct{}{}
	}
	for idx := range relativelyDeepest {
		for _, child := range getChildIndices(idx, sorted, indexToDepth) {
			allIndices[child] = struct{}{}
		}
	}

	// Expand by parents for all found indices.
	// Snapshot first to match Python's list(all_indices) behavior and avoid
	// mutating the map while ranging over it (undefined visit order in Go).
	toExpand := make([]int, 0, len(allIndices))
	for idx := range allIndices {
		toExpand = append(toExpand, idx)
	}
	for _, idx := range toExpand {
		for _, parent := range getParentIndices(idx, sorted, indexToDepth) {
			allIndices[parent] = struct{}{}
		}
	}

	// Collect in sorted order
	keys := make([]int, 0, len(allIndices))
	for idx := range allIndices {
		keys = append(keys, idx)
	}
	sort.Ints(keys)

	result := make([]ChunkInput, 0, len(keys))
	for _, idx := range keys {
		if ch, ok := indexToChunk[idx]; ok {
			result = append(result, ch)
		}
	}
	return result, nil
}

// isAncestor returns true if idx1 is a strict ancestor of idx2:
// idx1 < idx2, depth[idx1] < depth[idx2], and no chunk between them has depth ≤ depth[idx1].
func isAncestor(idx1, idx2 int, sorted []ChunkInput, indexToDepth map[int]int) bool {
	if idx1 >= idx2 {
		return false
	}
	d1, ok1 := indexToDepth[idx1]
	d2, ok2 := indexToDepth[idx2]
	if !ok1 || !ok2 {
		return false
	}
	if d1 >= d2 {
		return false
	}
	for _, ch := range sorted {
		if ch.ChunkIndex <= idx1 {
			continue
		}
		if ch.ChunkIndex > idx2 {
			break
		}
		if ch.ChunkIndex == idx2 {
			continue
		}
		if indexToDepth[ch.ChunkIndex] <= d1 {
			return false
		}
	}
	return true
}

// getChildIndices returns all chunk indices that are descendants of chunkIndex
// (deeper depth, no break in the chain at or above base depth).
func getChildIndices(chunkIndex int, sorted []ChunkInput, indexToDepth map[int]int) []int {
	baseDepth, ok := indexToDepth[chunkIndex]
	if !ok {
		return nil
	}
	var children []int
	past := false
	for _, ch := range sorted {
		if ch.ChunkIndex == chunkIndex {
			past = true
			continue
		}
		if !past {
			continue
		}
		if ch.Depth <= baseDepth {
			break
		}
		children = append(children, ch.ChunkIndex)
	}
	return children
}

// getParentIndices returns ancestor chunk indices in root→leaf order
// (walking backward, collecting strictly decreasing depths).
func getParentIndices(chunkIndex int, sorted []ChunkInput, indexToDepth map[int]int) []int {
	currentDepth, ok := indexToDepth[chunkIndex]
	if !ok {
		return nil
	}
	var parents []int
	// Find position of chunkIndex
	pos := -1
	for i, ch := range sorted {
		if ch.ChunkIndex == chunkIndex {
			pos = i
			break
		}
	}
	if pos < 0 {
		return nil
	}
	for i := pos - 1; i >= 0; i-- {
		if sorted[i].Depth < currentDepth {
			parents = append(parents, sorted[i].ChunkIndex)
			currentDepth = sorted[i].Depth
		}
	}
	// Reverse to get root→leaf order
	for l, r := 0, len(parents)-1; l < r; l, r = l+1, r-1 {
		parents[l], parents[r] = parents[r], parents[l]
	}
	return parents
}
