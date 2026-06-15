package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/ideaproject/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\n", strField(row, "id"), strField(row, "name"))
		}
		w.Flush()
		return nil
	},
}

var projectsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/ideaproject/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd, projectsShowCmd)
}
