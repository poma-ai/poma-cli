package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// CheatsheetCmd returns the cheatsheet command group.
func CheatsheetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cheatsheet",
		Short: "Generate cheatsheets from POMA chunk data (local, no API call)",
	}
	cmd.AddCommand(cheatsheetCreateCmd())
	return cmd
}

func cheatsheetCreateCmd() *cobra.Command {
	var input string
	var all bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate cheatsheet text from relevant_chunksets and all_chunks",
		Long: `Generate cheatsheet text from a JSON object containing "relevant_chunksets" and "all_chunks".

The --input value is either an inline JSON string (starting with '{') or a path to a .json file.

Default output: plain text content of the first cheatsheet.
With --all: JSON array of {file_id, content} objects for every document.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("--input is required")
			}

			var raw []byte
			if strings.HasPrefix(strings.TrimSpace(input), "{") {
				raw = []byte(input)
			} else {
				safePath, err := client.ValidateInputFilePath(input)
				if err != nil {
					return err
				}
				b, err := os.ReadFile(safePath)
				if err != nil {
					return fmt.Errorf("reading input file: %w", err)
				}
				raw = b
			}

			var req client.CheatsheetRequest
			if err := json.Unmarshal(raw, &req); err != nil {
				return fmt.Errorf("parsing input JSON: %w", err)
			}

			if len(req.RelevantChunksets) == 0 {
				return fmt.Errorf("input JSON must contain a non-empty 'relevant_chunksets' array")
			}
			if len(req.AllChunks) == 0 {
				return fmt.Errorf("input JSON must contain a non-empty 'all_chunks' array")
			}

			cheatsheets, err := client.GenerateCheatsheets(req.RelevantChunksets, req.AllChunks)
			if err != nil {
				return err
			}
			if len(cheatsheets) == 0 {
				return fmt.Errorf("no cheatsheet could be generated from the provided input")
			}

			if all {
				out, err := json.MarshalIndent(cheatsheets, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Println(cheatsheets[0].Content)
			return nil
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", `Inline JSON string or path to .json file (required)`)
	cmd.Flags().BoolVar(&all, "all", false, "Output all cheatsheets as JSON array instead of plain text")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}
