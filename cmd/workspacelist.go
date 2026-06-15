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
		w := newTabWriter()
		fmt.Fprintln(w, "SLUG\tNAME")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\n", strField(row, "slug"), strField(row, "name"))
		}
		w.Flush()
		return nil
	},
}
