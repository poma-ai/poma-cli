package cli

import (
	"fmt"

	"github.com/poma-ai/poma-cli/pkg/client"
	"github.com/spf13/cobra"
)

// OrgaCmd returns the orga command group.
func OrgaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orga",
		Short: "Organisation management",
	}
	cmd.AddCommand(
		orgaCreateCmd(),
		orgaGetCmd(),
		orgaUpdateCmd(),
		orgaDeleteCmd(),
		orgaMembersCmd(),
		orgaProjectsCmd(),
	)
	return cmd
}

func orgaCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create organisation (POST /orgas)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(name, "--name"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.CreateOrga(&client.CreateOrgaRequest{Name: name})
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Organisation name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func orgaGetCmd() *cobra.Command {
	var orgaID string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get organisation (GET /orgas/{orgaId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetOrga(orgaID)
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	return cmd
}

func orgaUpdateCmd() *cobra.Command {
	var orgaID, name string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update organisation (PUT /orgas/{orgaId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateResourceName(name, "--name"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.UpdateOrga(orgaID, &client.UpdateOrgaRequest{Name: name})
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "New organisation name (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func orgaDeleteCmd() *cobra.Command {
	var orgaID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete organisation (DELETE /orgas/{orgaId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.DeleteOrga(orgaID)
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	return cmd
}

// orgaMembersCmd returns the members subcommand group.
func orgaMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "Manage organisation members",
	}
	cmd.AddCommand(
		orgaMembersListCmd(),
		orgaMembersAddCmd(),
		orgaMembersUpdateRoleCmd(),
		orgaMembersRemoveCmd(),
	)
	return cmd
}

func orgaMembersListCmd() *cobra.Command {
	var orgaID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organisation members (GET /orgas/{orgaId}/members)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetOrgaMembers(orgaID)
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	return cmd
}

func orgaMembersAddCmd() *cobra.Command {
	var orgaID, accountID, role string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add member to organisation (POST /orgas/{orgaId}/members)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateResourceName(accountID, "--account-id"); err != nil {
				return err
			}
			// "owner" is not valid for new members; use members update-role to promote an existing member.
			if role != "admin" && role != "member" {
				return fmt.Errorf("--role must be 'admin' or 'member'")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.AddOrgaMember(orgaID, &client.AddOrgaMemberRequest{
				AccountID: accountID,
				Role:      role,
			})
			if err != nil {
				return err
			}
			if status != 201 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "Account ID to add (required)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Role: admin or member (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("account-id")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func orgaMembersUpdateRoleCmd() *cobra.Command {
	var orgaID, accountID, role string
	cmd := &cobra.Command{
		Use:   "update-role",
		Short: "Update member role (PUT /orgas/{orgaId}/members/{accountId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateResourceName(accountID, "--account-id"); err != nil {
				return err
			}
			if role != "owner" && role != "admin" && role != "member" {
				return fmt.Errorf("--role must be 'owner', 'admin', or 'member'")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.UpdateOrgaMemberRole(orgaID, accountID, &client.UpdateOrgaMemberRoleRequest{Role: role})
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "Account ID (required)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Role: owner, admin, or member (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("account-id")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func orgaMembersRemoveCmd() *cobra.Command {
	var orgaID, accountID string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove member from organisation (DELETE /orgas/{orgaId}/members/{accountId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateResourceName(accountID, "--account-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.RemoveOrgaMember(orgaID, accountID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			fmt.Println("removed")
			return nil
		},
	}
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "Account ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("account-id")
	return cmd
}

func orgaProjectsCmd() *cobra.Command {
	var orgaID string
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List projects for an organisation (GET /orgas/{orgaId}/projects)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.GetOrgaProjects(orgaID)
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
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	return cmd
}
