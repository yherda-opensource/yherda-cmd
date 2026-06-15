package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var workspaceListCmd = &cobra.Command{
	Use:   "workspacelist",
	Short: "List all workspaces available to the authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustPublicClient()
		var result []map[string]any
		if err := client.Get("/tenants/tenant/mine/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		for _, row := range result {
			fmt.Println(strField(row, "name"))
		}
		return nil
	},
}
