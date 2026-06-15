package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Yherda via OAuth2",
	Long:  "Opens a browser to authenticate with Yherda. Stores the token in ~/.yherdacmd/credentials.json.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// OAuth2 PKCE flow — to be implemented when GEN-162 (Django OAuth provider) is complete.
		// Placeholder: print instructions.
		fmt.Println("OAuth2 login not yet implemented — requires GEN-162 (backend OAuth provider).")
		fmt.Println("Once implemented, this command will open a browser to authenticate with Yherda.")
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored Yherda credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.DeleteCredentials(); err != nil {
			return fmt.Errorf("failed to remove credentials: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}
