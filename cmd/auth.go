package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/auth"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var loginIncludeIdeaOwnerRead bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Yherda via OAuth2",
	Long: `Opens a browser to authenticate with Yherda using OAuth2 PKCE. Once you
approve the login, the resulting token is stored in
~/.yherdacmd/credentials.json and used automatically (and refreshed
automatically) by every other command.

By default, requests read-only access to Ideas you own in addition to
Collaborator (Occupation-reachable) read/write access. Use
--idea-owner-read=false to request Collaborator scope only.

Run this once per machine before doing anything else with the CLI.`,
	Example: `  yherda login
  yherda login --idea-owner-read=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Login(loginIncludeIdeaOwnerRead)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}
		fmt.Println("Logged in successfully. Credentials saved to ~/.yherdacmd/credentials.json")
		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&loginIncludeIdeaOwnerRead, "idea-owner-read", true,
		"Request read-only access to Ideas you own, in addition to Collaborator scope (disable with --idea-owner-read=false for a Collaborator-only login)")
}

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Remove stored Yherda credentials",
	Long:    "Deletes ~/.yherdacmd/credentials.json. You'll need to 'yherda login' again before running any other command.",
	Example: `  yherda logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.DeleteCredentials(); err != nil {
			return fmt.Errorf("failed to remove credentials: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}
