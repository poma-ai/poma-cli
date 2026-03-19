package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// HealthCmd returns the health command.
func HealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Service health GET /health",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			body, status, err := cli.Health()
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
