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
		orgaListCmd(),
		orgaCreateCmd(),
		orgaGetCmd(),
		orgaUpdateCmd(),
		orgaDeleteCmd(),
		orgaMembersCmd(),
		orgaProjectsCmd(),
		orgaInvitationsCmd(),
		orgaAcceptInvitationCmd(),
	)
	return cmd
}

func orgaListCmd() *cobra.Command {
	var name string
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List my organisations (GET /orgas)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name != "" {
				if err := client.ValidateUserStrings(name, "", "", ""); err != nil {
					return err
				}
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.ListOrgas(name, page, pageSize)
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
	cmd.Flags().StringVarP(&name, "name", "n", "", "Filter by organisation name (substring match)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (1-based)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Results per page")
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
	var orgaID, email, role string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add member to organisation by email (POST /orgas/{orgaId}/members)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateUserStrings(email, "", "", ""); err != nil {
				return err
			}
			if role != "admin" && role != "member" {
				return fmt.Errorf("--role must be 'admin' or 'member'")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.AddOrgaMember(orgaID, &client.AddOrgaMemberRequest{
				Email: email,
				Role:  role,
			})
			if err != nil {
				return err
			}
			if status != 201 && status != 202 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address of account to add (required)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Role: admin or member (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("email")
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

// orgaInvitationsCmd returns the invitations subcommand group.
func orgaInvitationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitations",
		Short: "Manage organisation invitations",
	}
	cmd.AddCommand(
		orgaInvitationsInviteCmd(),
		orgaInvitationsListCmd(),
		orgaInvitationsCancelCmd(),
		orgaInvitationsResendCmd(),
	)
	return cmd
}

func orgaInvitationsInviteCmd() *cobra.Command {
	var orgaID, email string
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite account to organisation (POST /orgas/{orgaId}/invitations)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if err := client.ValidateUserStrings(email, "", "", ""); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.CreateOrgaInvitation(orgaID, &client.CreateOrgaInvitationRequest{Email: email})
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
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address to invite (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func orgaInvitationsListCmd() *cobra.Command {
	var orgaID, status string
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organisation invitations (GET /orgas/{orgaId}/invitations)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, st, err := cli.ListOrgaInvitations(orgaID, status, page, pageSize)
			if err != nil {
				return err
			}
			if st != 200 {
				return fmt.Errorf("HTTP %d: %s", st, string(body))
			}
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status: pending, accepted, cancelled, expired, all (default: pending)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (1-based)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Results per page")
	_ = cmd.MarkFlagRequired("orga-id")
	return cmd
}

func orgaInvitationsCancelCmd() *cobra.Command {
	var orgaID string
	var invitationID int64
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel organisation invitation (DELETE /orgas/{orgaId}/invitations/{invitationId})",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if invitationID <= 0 {
				return fmt.Errorf("--invitation-id must be a positive integer")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.CancelOrgaInvitation(orgaID, invitationID)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			fmt.Println("cancelled")
			return nil
		},
	}
	cmd.Flags().StringVarP(&orgaID, "orga-id", "o", "", "Organisation ID (required)")
	cmd.Flags().Int64VarP(&invitationID, "invitation-id", "i", 0, "Invitation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("invitation-id")
	return cmd
}

func orgaInvitationsResendCmd() *cobra.Command {
	var orgaID string
	var invitationID int64
	cmd := &cobra.Command{
		Use:   "resend",
		Short: "Resend organisation invitation email (POST /orgas/{orgaId}/invitations/{invitationId}/resend)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateResourceName(orgaID, "--orga-id"); err != nil {
				return err
			}
			if invitationID <= 0 {
				return fmt.Errorf("--invitation-id must be a positive integer")
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.ResendOrgaInvitation(orgaID, invitationID)
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
	cmd.Flags().Int64VarP(&invitationID, "invitation-id", "i", 0, "Invitation ID (required)")
	_ = cmd.MarkFlagRequired("orga-id")
	_ = cmd.MarkFlagRequired("invitation-id")
	return cmd
}

func orgaAcceptInvitationCmd() *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "accept-invitation",
		Short: "Accept organisation invitation (GET /invitations/accept)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.ValidateUserStrings(token, "", "", ""); err != nil {
				return err
			}
			cli := apiClient()
			if err := requireToken(cli.Token); err != nil {
				return err
			}
			body, status, err := cli.AcceptOrgaInvitation(token)
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
	cmd.Flags().StringVarP(&token, "token", "t", "", "One-time invitation token from the invitation email (required)")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

