package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// JobsCmd returns the jobs command.
func JobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Job ingestion, status, download, and delete",
	}
	cmd.AddCommand(
		ingestCmd(),
		ingestDataCmd(),
		ingestEcoCmd(),
		ingestEcoDataCmd(),
		jobStatusCmd(),
		jobStatusStreamCmd(),
		jobDownloadCmd(),
		jobDeleteCmd(),
	)
	return cmd
}

func ingestCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest file (raw body, pro) POST /ingest",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateIngestFilePath(file); err != nil {
				return err
			}
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.Ingest(file)
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			return printIngestJobIDOnly(body)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func ingestDataCmd() *cobra.Command {
	var data, filename string
	cmd := &cobra.Command{
		Use:   "ingest-data",
		Short: "Ingest raw bytes (pro) POST /ingest",
		Long: "Send the request body from --data or from stdin (pipe a file: poma jobs ingest-data < doc.pdf). " +
			"For binary files prefer stdin; --data is best for small text payloads. " +
			"--filename sets the Content-Disposition basename (required).",
		RunE: func(cmd *cobra.Command, args []string) error {
			var payload []byte
			var err error
			if filename == "" {
				return fmt.Errorf("filename is required for ingest-data")
			}
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
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.IngestData(payload, filename)
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			return printIngestJobIDOnly(body)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Inline body (use stdin for binary or large content)")
	cmd.Flags().StringVarP(&filename, "filename", "f", "", `Basename for Content-Disposition`)
	return cmd
}

func ingestEcoCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "ingest-eco",
		Short: "Ingest file (raw body, eco) POST /ingestEco",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateIngestFilePath(file); err != nil {
				return err
			}
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.IngestEco(file)
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			return printIngestJobIDOnly(body)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file to ingest")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func ingestEcoDataCmd() *cobra.Command {
	var data, filename string
	cmd := &cobra.Command{
		Use:   "ingest-eco-data",
		Short: "Ingest raw bytes (eco) POST /ingestEco",
		Long: "Send the request body from --data or from stdin (pipe a file: poma jobs ingest-eco-data < doc.pdf). " +
			"For binary files prefer stdin; --data is best for small text payloads. " +
			"--filename sets the Content-Disposition basename (required).",
		RunE: func(cmd *cobra.Command, args []string) error {
			var payload []byte
			var err error
			if filename == "" {
				return fmt.Errorf("filename is required for ingest-eco-data")
			}
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
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.IngestEcoData(payload, filename)
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			return printIngestJobIDOnly(body)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Inline body (use stdin for binary or large content)")
	cmd.Flags().StringVarP(&filename, "filename", "n", "", `Basename for Content-Disposition`)
	return cmd
}

// printIngestJobIDOnly writes pretty-printed {"job_id":"..."} to stdout (normalized via ParseJob).
func printIngestJobIDOnly(body []byte) error {
	j, err := client.ParseJob(body)
	if err != nil {
		return fmt.Errorf("parse ingest response: %w", err)
	}
	if j.JobID == "" {
		return fmt.Errorf("ingest response has no job_id")
	}
	out, err := json.MarshalIndent(map[string]string{"job_id": j.JobID}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

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
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
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
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			ctx := context.Background()
			if err := cli.StatusStream(ctx, jobID, statusBaseURLOrDefault(), func(s *client.JobStatus) bool {
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

func jobDownloadCmd() *cobra.Command {
	var jobID, output string
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download job result GET /jobs/{job_id}/download",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateJobID(jobID); err != nil {
				return err
			}
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			var safeOut string
			if output != "" {
				var err error
				safeOut, err = client.ValidateSafeOutputDir(output)
				if err != nil {
					return err
				}
			} else {
				var err error
				safeOut, err = client.ValidateSafeOutputDir(filepath.Join("bin", client.PomaArchiveName(jobID)))
				if err != nil {
					return err
				}
			}
			n, status, err := cli.DownloadJob(jobID, safeOut)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d", status)
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
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
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
