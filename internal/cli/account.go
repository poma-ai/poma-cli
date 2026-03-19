package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AccountCmd returns the account command.
func AccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Account info and usage",
	}
	cmd.AddCommand(meCmd(), myProjectsCmd(), myUsageCmd())
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
