package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// PrimeCutCmd returns the primecut command.
func PrimeCutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "primecut",
		Short: "PrimeCut ingest, and ingest-sync",
	}
	cmd.AddCommand(
		primecutIngestCmd(),
		primecutIngestSyncCmd(),
	)
	return cmd
}

func primecutIngestCmd() *cobra.Command {
	var file, data, filename string
	var eco bool
	cmd := &cobra.Command{
		Use:     "ingest",
		Aliases: []string{"ingest-data", "ingest-eco", "ingest-eco-data"},
		Short:   "Ingest POST /ingest or /ingestEco — from a file path or from stdin / --data",
		Long: "Either pass a path with --file / -f, or send the raw body with --filename / -n " +
			"and --data or stdin (e.g. poma job ingest -n doc.pdf < doc.pdf). " +
			"Use --eco or the ingest-eco / ingest-eco-data aliases for POST /ingestEco. " +
			"Do not combine --file with --data. Binary payloads should use stdin, not --data.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			isEco := eco
			switch cmd.CalledAs() {
			case "ingest-eco", "ingest-eco-data":
				isEco = true
			}

			if file != "" {
				if data != "" {
					return fmt.Errorf("use either --file or --data, not both")
				}
				if err := client.ValidateIngestFilePath(file); err != nil {
					return err
				}
				var body []byte
				var status int
				var err error
				if isEco {
					body, status, err = cli.IngestEco(file)
				} else {
					body, status, err = cli.Ingest(file)
				}
				if err != nil {
					return err
				}
				if status != 201 {
					return fmt.Errorf("HTTP %d: %s", status, string(body))
				}
				return client.PrintIngestJobIDOnly(body)
			}

			if filename == "" {
				return fmt.Errorf("without --file, --filename / -n is required (Content-Disposition basename)")
			}
			var payload []byte
			var err error
			if data != "" {
				payload = []byte(data)
			} else {
				payload, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
			}
			if len(payload) == 0 {
				return fmt.Errorf("no ingest payload: set --data or pipe bytes to stdin")
			}
			var body []byte
			var status int
			if isEco {
				body, status, err = cli.IngestEcoData(payload, filename)
			} else {
				body, status, err = cli.IngestData(payload, filename)
			}
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			return client.PrintIngestJobIDOnly(body)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest (mutually exclusive with stdin/--data body mode)")
	cmd.Flags().StringVar(&data, "data", "", "Inline body when not using --file (prefer stdin for binary)")
	cmd.Flags().StringVarP(&filename, "filename", "n", "", "Basename for Content-Disposition when not using --file (required in that mode)")
	cmd.Flags().BoolVar(&eco, "eco", false, "Use POST /ingestEco instead of /ingest")
	return cmd
}

func primecutIngestSyncCmd() *cobra.Command {
	var file, data, filename, output string
	var eco bool
	cmd := &cobra.Command{
		Use:   "ingest-sync",
		Short: "Ingest (pro or eco), stream status until terminal, then either receive json or download archive if done",
		Long: "POST /ingest (default) or POST /ingestEco with --eco, then subscribe to the status SSE stream (same as status-stream), " +
			"then GET /jobs/{job_id}/download when the job reaches done. " +
			"Use --file / -f for a path, or --filename / -n with --data or stdin (same as job ingest). " +
			"Do not combine --file with --data. " +
			"--output / -o to download the archive (default: {job_id}.poma) and print the path, otherwise receive json on stdout. " +
			"If the terminal status is failed or deleted, exits non-zero. Each status event is printed as JSON on stdout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			var resolve func(jobID string) (string, error)
			if output != "" {
				out := output
				resolve = func(jobID string) (string, error) {
					return resolveJobDownloadPath(jobID, out)
				}
			}
			onStatus := func(s *client.JobStatus) {
				b, _ := json.Marshal(s)
				PrintJSON(b)
			}
			ctx := cmd.Context()
			statusURL := statusBaseURLOrDefault()

			finish := func(n int64, jobIDOrPath string, err error) error {
				if err != nil {
					return err
				}
				if output == "" {
					body, status, err := cli.GetJobResult(jobIDOrPath)
					if err != nil {
						return err
					}
					if status != 200 {
						return fmt.Errorf("HTTP %d: %s", status, string(body))
					}
					PrintJSON(body)
					return nil
				}
				fmt.Printf("Downloaded %d bytes to %s\n", n, jobIDOrPath)
				return nil
			}

			if file != "" {
				if data != "" {
					return fmt.Errorf("use either --file or --data, not both")
				}
				if err := client.ValidateIngestFilePath(file); err != nil {
					return err
				}
				n, out, err := cli.IngestSync(ctx, file, eco, statusURL, resolve, onStatus)
				return finish(n, out, err)
			}

			if filename == "" {
				return fmt.Errorf("without --file, --filename / -n is required (Content-Disposition basename)")
			}
			var payload []byte
			var err error
			if data != "" {
				payload = []byte(data)
			} else {
				payload, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
			}
			if len(payload) == 0 {
				return fmt.Errorf("no ingest payload: set --data or pipe bytes to stdin")
			}
			n, out, err := cli.IngestDataSync(ctx, payload, filename, eco, statusURL, resolve, onStatus)
			return finish(n, out, err)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest (mutually exclusive with stdin/--data body mode)")
	cmd.Flags().StringVar(&data, "data", "", "Inline body when not using --file (prefer stdin for binary)")
	cmd.Flags().StringVarP(&filename, "filename", "n", "", "Basename for Content-Disposition when not using --file (required in that mode)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Download path (default: bin/{job_id}.poma)")
	cmd.Flags().BoolVar(&eco, "eco", false, "Use POST /ingestEco instead of /ingest")
	return cmd
}
