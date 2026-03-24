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
		Short: "Registration, verification, and account info",
	}
	cmd.AddCommand(
		registerEmailCmd(),
		verifyEmailCmd(),
		meCmd(),
		apiKeyCmd(),
		myProjectsCmd(),
		myUsageCmd(),
	)
	return cmd
}

func registerEmailCmd() *cobra.Command {
	var email, username, company string
	cmd := &cobra.Command{
		Use:   "register-email",
		Short: "Register a new user or login by email (POST /registerEmail)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				return fmt.Errorf("email is required")
			}
			if err := client.ValidateUserStrings(email, username, company, ""); err != nil {
				return err
			}
			req := &client.AccountRegisterEmailRequest{
				Email:    email,
				Username: username,
				Company:  company,
			}
			cli := apiClient()
			body, status, err := cli.RegisterEmail(req)
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
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email (required)")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	cmd.Flags().StringVarP(&company, "company", "c", "", "Company")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func verifyEmailCmd() *cobra.Command {
	var email, code string
	cmd := &cobra.Command{
		Use:   "verify-email",
		Short: "Verify email with code and get JWT (POST /verifyEmail)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" || code == "" {
				return fmt.Errorf("email and code are required")
			}
			if err := client.ValidateUserStrings(email, "", "", code); err != nil {
				return err
			}
			req := &client.AccountVerifyEmailRequest{
				Email: email,
				Code:  code,
			}
			cli := apiClient()
			body, status, err := cli.VerifyEmail(req)
			if err != nil {
				return err
			}
			if status != 200 {
				return fmt.Errorf("HTTP %d: %s", status, string(body))
			}
			var resp client.AccountRegisterEmailResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				PrintJSON(body)
				return nil
			}
			fmt.Println("Token:", resp.Token)
			PrintJSON(body)
			return nil
		},
	}
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email (required)")
	cmd.Flags().StringVarP(&code, "code", "k", "", "Verification code (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("code")
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
		Use:     "api-key",
		Aliases: []string{"apikey"},
		Short:   "Get long-lived API key GET /me (api_key only)",
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
				return fmt.Errorf("parse /me: %w", err)
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
