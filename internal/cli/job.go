package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// JobCmd returns the job command.
func JobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Job status, result, download, and delete",
	}
	cmd.AddCommand(
		// ingestCmd(),
		// ingestSyncCmd(),
		jobStatusCmd(),
		jobStatusStreamCmd(),
		jobResultCmd(),
		jobDownloadCmd(),
		jobDeleteCmd(),
	)
	return cmd
}

// func ingestCmd() *cobra.Command {
// 	var file, data, filename string
// 	var eco bool
// 	cmd := &cobra.Command{
// 		Use:     "ingest",
// 		Aliases: []string{"ingest-data", "ingest-eco", "ingest-eco-data"},
// 		Short:   "Ingest POST /ingest or /ingestEco — from a file path or from stdin / --data",
// 		Long: "Either pass a path with --file / -f, or send the raw body with --filename / -n " +
// 			"and --data or stdin (e.g. poma job ingest -n doc.pdf < doc.pdf). " +
// 			"Use --eco or the ingest-eco / ingest-eco-data aliases for POST /ingestEco. " +
// 			"Do not combine --file with --data. Binary payloads should use stdin, not --data.",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cli := apiClient()
// 			if err := requireToken(cli.Token); err != nil {
// 				return err
// 			}
// 			isEco := eco
// 			switch cmd.CalledAs() {
// 			case "ingest-eco", "ingest-eco-data":
// 				isEco = true
// 			}

// 			if file != "" {
// 				if data != "" {
// 					return fmt.Errorf("use either --file or --data, not both")
// 				}
// 				if err := client.ValidateIngestFilePath(file); err != nil {
// 					return err
// 				}
// 				var body []byte
// 				var status int
// 				var err error
// 				if isEco {
// 					body, status, err = cli.IngestEco(file)
// 				} else {
// 					body, status, err = cli.Ingest(file)
// 				}
// 				if err != nil {
// 					return err
// 				}
// 				if status != 201 {
// 					return fmt.Errorf("HTTP %d: %s", status, string(body))
// 				}
// 				return client.PrintIngestJobIDOnly(body)
// 			}

// 			if filename == "" {
// 				return fmt.Errorf("without --file, --filename / -n is required (Content-Disposition basename)")
// 			}
// 			var payload []byte
// 			var err error
// 			if data != "" {
// 				payload = []byte(data)
// 			} else {
// 				payload, err = io.ReadAll(cmd.InOrStdin())
// 				if err != nil {
// 					return err
// 				}
// 			}
// 			if len(payload) == 0 {
// 				return fmt.Errorf("no ingest payload: set --data or pipe bytes to stdin")
// 			}
// 			var body []byte
// 			var status int
// 			if isEco {
// 				body, status, err = cli.IngestEcoData(payload, filename)
// 			} else {
// 				body, status, err = cli.IngestData(payload, filename)
// 			}
// 			if err != nil {
// 				return err
// 			}
// 			if status != 201 {
// 				return fmt.Errorf("HTTP %d: %s", status, string(body))
// 			}
// 			return client.PrintIngestJobIDOnly(body)
// 		},
// 	}
// 	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest (mutually exclusive with stdin/--data body mode)")
// 	cmd.Flags().StringVar(&data, "data", "", "Inline body when not using --file (prefer stdin for binary)")
// 	cmd.Flags().StringVarP(&filename, "filename", "n", "", "Basename for Content-Disposition when not using --file (required in that mode)")
// 	cmd.Flags().BoolVar(&eco, "eco", false, "Use POST /ingestEco instead of /ingest")
// 	return cmd
// }

// func ingestSyncCmd() *cobra.Command {
// 	var file, data, filename, output string
// 	var eco bool
// 	cmd := &cobra.Command{
// 		Use:   "ingest-sync",
// 		Short: "Ingest (pro or eco), stream status until terminal, then either receive json or download archive if done",
// 		Long: "POST /ingest (default) or POST /ingestEco with --eco, then subscribe to the status SSE stream (same as status-stream), " +
// 			"then GET /jobs/{job_id}/download when the job reaches done. " +
// 			"Use --file / -f for a path, or --filename / -n with --data or stdin (same as job ingest). " +
// 			"Do not combine --file with --data. " +
// 			"--output / -o to download the archive (default: {job_id}.poma) and print the path, otherwise receive json on stdout. " +
// 			"If the terminal status is failed or deleted, exits non-zero. Each status event is printed as JSON on stdout.",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cli := apiClient()
// 			if err := requireToken(cli.Token); err != nil {
// 				return err
// 			}
// 			var resolve func(jobID string) (string, error)
// 			if output != "" {
// 				out := output
// 				resolve = func(jobID string) (string, error) {
// 					return resolveJobDownloadPath(jobID, out)
// 				}
// 			}
// 			onStatus := func(s *client.JobStatus) {
// 				b, _ := json.Marshal(s)
// 				PrintJSON(b)
// 			}
// 			ctx := cmd.Context()
// 			statusURL := statusBaseURLOrDefault()

// 			finish := func(n int64, jobIDOrPath string, err error) error {
// 				if err != nil {
// 					return err
// 				}
// 				if output == "" {
// 					body, status, err := cli.GetJobResult(jobIDOrPath)
// 					if err != nil {
// 						return err
// 					}
// 					if status != 200 {
// 						return fmt.Errorf("HTTP %d: %s", status, string(body))
// 					}
// 					PrintJSON(body)
// 					return nil
// 				}
// 				fmt.Printf("Downloaded %d bytes to %s\n", n, jobIDOrPath)
// 				return nil
// 			}

// 			if file != "" {
// 				if data != "" {
// 					return fmt.Errorf("use either --file or --data, not both")
// 				}
// 				if err := client.ValidateIngestFilePath(file); err != nil {
// 					return err
// 				}
// 				n, out, err := cli.IngestSync(ctx, file, eco, statusURL, resolve, onStatus)
// 				return finish(n, out, err)
// 			}

// 			if filename == "" {
// 				return fmt.Errorf("without --file, --filename / -n is required (Content-Disposition basename)")
// 			}
// 			var payload []byte
// 			var err error
// 			if data != "" {
// 				payload = []byte(data)
// 			} else {
// 				payload, err = io.ReadAll(cmd.InOrStdin())
// 				if err != nil {
// 					return err
// 				}
// 			}
// 			if len(payload) == 0 {
// 				return fmt.Errorf("no ingest payload: set --data or pipe bytes to stdin")
// 			}
// 			n, out, err := cli.IngestDataSync(ctx, payload, filename, eco, statusURL, resolve, onStatus)
// 			return finish(n, out, err)
// 		},
// 	}
// 	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest (mutually exclusive with stdin/--data body mode)")
// 	cmd.Flags().StringVar(&data, "data", "", "Inline body when not using --file (prefer stdin for binary)")
// 	cmd.Flags().StringVarP(&filename, "filename", "n", "", "Basename for Content-Disposition when not using --file (required in that mode)")
// 	cmd.Flags().StringVarP(&output, "output", "o", "", "Download path (default: bin/{job_id}.poma)")
// 	cmd.Flags().BoolVar(&eco, "eco", false, "Use POST /ingestEco instead of /ingest")
// 	return cmd
// }

func jobStatusCmd() *cobra.Command {
	var jobID string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get job status GET /jobs/{job_id}/status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetJobStatus(jobID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "Job ID")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func jobStatusStreamCmd() *cobra.Command {
	var jobID string
	cmd := &cobra.Command{
		Use:   "status-stream",
		Short: "Stream job status via SSE until terminal state (GET status/v1/jobs/{job_id})",
		Long:  "Subscribe to the Status API SSE stream for a job. Prints each status event until the job reaches done, failed, or deleted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			if err := cli.StatusStream(cmd.Context(), jobID, statusBaseURLOrDefault(), func(s *client.JobStatus) bool {
				data, _ := json.Marshal(s)
				PrintJSON(data)
				return true
			}); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "Job ID")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func jobResultCmd() *cobra.Command {
	var jobID string
	cmd := &cobra.Command{
		Use:   "result",
		Short: "Get job result JSON GET /jobs/{job_id}/results",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetJobResult(jobID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "Job ID")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func jobDownloadCmd() *cobra.Command {
	var jobID, output string
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download job archive to output.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			safeOut, err := resolveJobDownloadPath(jobID, output)
			if err != nil {
				return err
			}
			n, _, err := cli.DownloadJob(jobID, safeOut)
			if err != nil {
				return err
			}
			fmt.Printf("Downloaded %d bytes to %s\n", n, safeOut)
			return nil
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "Job ID")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path (default: bin/{job_id}.poma)")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func resolveJobDownloadPath(jobID, output string) (string, error) {
	if output != "" {
		return client.ValidateSafeOutputDir(output)
	}
	return client.ValidateSafeOutputDir(filepath.Join("bin", client.PomaArchiveName(jobID)))
}

func jobDeleteCmd() *cobra.Command {
	var jobID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a job (best-effort) DELETE /jobs/{job_id}",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.DeleteJob(jobID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			fmt.Println("Delete accepted (best-effort)")
			return nil
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "Job ID")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}
