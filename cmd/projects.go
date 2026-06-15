package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Projects are expressions scoped to the reserved "_projects" idea.
// The CLI finds the _projects idea first, then operates on its expressions.

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects (expressions in the _projects idea)",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		// Find the _projects idea, then list its expressions.
		var ideas []map[string]any
		if err := client.Get("/storyline/?name=_projects", &ideas); err != nil {
			return err
		}
		if len(ideas) == 0 {
			return fmt.Errorf("no _projects idea found in this workspace")
		}
		ideaID := fmt.Sprintf("%v", ideas[0]["id"])
		var result any
		if err := client.Get("/storyline/"+ideaID+"/expressions/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
}
