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
	Long:  "Opens a browser to authenticate with Yherda. Stores the token in ~/.yherdacmd/credentials.json.",
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
