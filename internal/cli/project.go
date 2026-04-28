package cli

import (
	"fmt"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// ProjectCmd returns the project command group.
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management",
	}
	cmd.AddCommand(
		projectCreateCmd(),
		projectListCmd(),
		projectSearchCmd(),
		projectGetCmd(),
		projectDeleteCmd(),
	)
	return cmd
}

func projectCreateCmd() *cobra.Command {
	var name, product, accountID, orgaID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create project (POST /projects)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(name, "--name"); err != nil {
				return err
			}
			if product != "primecut" && product != "grill" {
				return fmt.Errorf("--product must be 'primecut' or 'grill'")
			}
			if accountID != "" {
				if err := client.ValidateResourceName(accountID, "--account-id"); err != nil {
					return err
				}
			}
			if orgaID != "" {
				if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
					return err
				}
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.CreateProject(&client.CreateProjectRequest{
				Product:   product,
				Name:      name,
				AccountID: accountID,
				OrgaID:    orgaID,
			})
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
	cmd.Flags().StringVarP(&name, "name", "n", "", "Project display name (required)")
	cmd.Flags().StringVarP(&product, "product", "p", "", "Product: primecut or grill (required)")
	cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "Account ID (defaults to authenticated account)")
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (creates an org-scoped project)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("product")
	return cmd
}

func projectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List my projects (GET /projects)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.ListProjects()
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

func projectSearchCmd() *cobra.Command {
	var accountID, orgaID, projectID, name, product string
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search projects (GET /projects/search)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountID != "" {
				if err := client.ValidateResourceName(accountID, "--account-id"); err != nil {
					return err
				}
			}
			if orgaID != "" {
				if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
					return err
				}
			}
			if projectID != "" {
				if err := client.ValidateResourceName(projectID, "--project-id"); err != nil {
					return err
				}
			}
			if product != "" && product != "primecut" && product != "grill" {
				return fmt.Errorf("--product must be 'primecut' or 'grill'")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.SearchProjects(&client.ProjectSearchOptions{
				AccountID: accountID,
				OrgaID:    orgaID,
				ProjectID: projectID,
				Name:      name,
				Product:   product,
			})
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
	cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "Filter by account ID")
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Filter by organisation ID")
	cmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Filter by project ID (slug)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Filter by project name")
	cmd.Flags().StringVarP(&product, "product", "P", "", "Filter by product: primecut or grill")
	return cmd
}

func projectGetCmd() *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get project by internal UUID (GET /projects/{projectId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(projectID, "--project-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetProject(projectID)
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
	cmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Internal project UUID (id field) (required)")
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}

func projectDeleteCmd() *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete project (DELETE /projects/{projectId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(projectID, "--project-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.DeleteProject(projectID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			fmt.Println("deleted")
			return nil
		},
	}
	cmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Internal project UUID (id field) (required)")
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}
