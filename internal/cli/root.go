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
			"Optional --json accepts inline JSON or a path to a JSON file; flag values override the file/JSON.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(jsonArg) != "" {
				cfg, err := parseFileConfig(jsonArg)
				if err != nil {
					return err
				}
				if err := validate_file_config(cfg); err != nil {
					return err
				}
				if err := mergeConfigIntoFlags(cmd, cfg); err != nil {
					return err
				}
			}
			return validate_persistent_flags()
		},
	}
	cmd.PersistentFlags().StringVar(&baseURL, "base-url", defaultApiBaseURL, "API base URL")
	cmd.PersistentFlags().StringVar(&statusBaseURL, "status-base-url", defaultStatusBaseURL, "Status SSE API base URL")
	cmd.PersistentFlags().StringVar(&token, "token", os.Getenv("POMA_API_TOKEN"), "JWT token (or set POMA_API_TOKEN)")
	cmd.PersistentFlags().StringVar(&jsonArg, "json", "", "JSON options (inline object or path to .json); explicit flags override")

	cmd.AddCommand(
		UserCmd(),
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
