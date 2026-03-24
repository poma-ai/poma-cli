package cli

import (
	"os"
	"strings"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

var (
	baseURL       string
	statusBaseURL string
	token         string
	jsonArg string
)

const (
	defaultApiBaseURL    = "https://api.poma-ai.com/v2"
	defaultStatusBaseURL = "https://api.poma-ai.com/status/v1"
)

// RootCmd returns the root command for the POMA CLI with all subcommands attached.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poma",
		Short: "POMA AI API v2 CLI",
		Long: "CLI for the POMA AI public API. Use --base-url and --token or POMA_API_TOKEN.\n" +
			"Optional --json accepts inline JSON or a path to a JSON file; flag values override the file/JSON.\n\n" +
			"Top-level commands: account, jobs, health. Subcommands (e.g. account register-email, account api-key, jobs ingest) are listed under each: poma <cmd> --help.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Do not use os.Getenv as the flag default — it would print the JWT in --help.
			if flg := cmd.Flags().Lookup("token"); flg != nil && !flg.Changed {
				token = strings.TrimSpace(os.Getenv("POMA_API_TOKEN"))
			}
			if strings.TrimSpace(jsonArg) != "" {
				cfg, err := parseFileConfig(jsonArg)
				if err != nil {
					return err
				}
				if err := client.ValidateFileConfig(cfg); err != nil {
					return err
				}
				if err := mergeConfigIntoFlags(cmd, cfg); err != nil {
					return err
				}
			}
			return client.ValidatePersistentFlags(baseURL, statusBaseURL, token, jsonArg)
		},
	}
	cmd.PersistentFlags().StringVar(&baseURL, "base-url", defaultApiBaseURL, "API base URL")
	cmd.PersistentFlags().StringVar(&statusBaseURL, "status-base-url", defaultStatusBaseURL, "Status SSE API base URL")
	cmd.PersistentFlags().StringVar(&token, "token", "", "JWT token (or set POMA_API_TOKEN)")
	cmd.PersistentFlags().StringVar(&jsonArg, "json", "", "JSON options (inline object or path to .json); explicit flags override")

	cmd.AddCommand(
		AccountCmd(),
		JobsCmd(),
		HealthCmd(),
	)
	return cmd
}

// Execute runs the root command. Call from main.
func Execute() error {
	return RootCmd().Execute()
}

func apiClient() *client.Client {
	return client.New(baseURL, token)
}

func statusBaseURLOrDefault() string {
	if statusBaseURL != "" {
		return statusBaseURL
	}
	return defaultStatusBaseURL
}
