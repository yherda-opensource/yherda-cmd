package cmd

import (
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
		var result any
		if err := client.Get("/ideaproject/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var projectsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result any
		if err := client.Get("/ideaproject/"+args[0]+"/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd, projectsShowCmd)
}
