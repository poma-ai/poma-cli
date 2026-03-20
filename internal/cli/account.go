package cli

import (
	"encoding/json"
	"fmt"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// AccountCmd returns the account command.
func AccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Account info and usage",
	}
	cmd.AddCommand(meCmd(), apiKeyCmd(), myProjectsCmd(), myUsageCmd())
	return cmd
}

func meCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get current account GET /me",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.GetMe()
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
	return cmd
}

func apiKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Get long-lived API key GET /accounts/me (api_key only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.GetAccountsMe()
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			var parsed client.AccountAPIKeyBody
			if err := json.Unmarshal(body, &parsed); err != nil {
				return fmt.Errorf("parse /accounts/me: %w", err)
			}
			if parsed.APIKey == "" {
				return fmt.Errorf("response has no api_key")
			}
			raw, err := json.Marshal(map[string]string{"api_key": parsed.APIKey})
			if err != nil {
				return err
			}
			PrintJSON(raw)
			return nil
		},
	}
	return cmd
}

func myProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "my-projects",
		Short: "List my projects GET /myProjects",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.GetMyProjects()
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
	return cmd
}

func myUsageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "my-usage",
		Short: "Get my usage GET /myUsage",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if cli.Token == "" {
				return fmt.Errorf("token is required (--token or POMA_API_TOKEN)")
			}
			body, status, err := cli.GetMyUsage()
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
	return cmd
}
