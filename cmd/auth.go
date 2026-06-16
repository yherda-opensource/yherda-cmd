package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/auth"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Yherda via OAuth2",
	Long: `Opens a browser to authenticate with Yherda using OAuth2 PKCE. Once you
approve the login, the resulting token is stored in
~/.yherdacmd/credentials.json and used automatically (and refreshed
automatically) by every other command.

Run this once per machine before doing anything else with the CLI.`,
	Example: `  yherda login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Login()
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
