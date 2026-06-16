package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace [name]",
	Short: "Show or set the active workspace",
	Long: `Show the active workspace (no argument) or set it by name.

The workspace name and its API server are looked up from your account and
stored in .yherda in the current directory. All subsequent API calls route to
the workspace's API server.

Switching workspaces clears the active idea/person/arc/place/thing context,
since those are scoped to a workspace. Run 'yherda workspacelist' to see
which workspaces are available to you.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  yherda workspace
  yherda workspace acme-publishing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return fmt.Errorf("failed to load context: %w", err)
		}

		if len(args) == 0 {
			if ctx.Workspace == "" {
				fmt.Println("No active workspace set. Run: yherda workspace <name>")
			} else {
				fmt.Printf("Active workspace: %s\nAPI Endpoint:     %s\n", ctx.Workspace, ctx.APIServer)
			}
			return nil
		}

		name := args[0]
		client := mustPublicClient()
		var workspaces []map[string]any
		if err := client.Get("/tenants/tenant/mine/", &workspaces); err != nil {
			return fmt.Errorf("failed to fetch workspaces: %w", err)
		}

		for _, ws := range workspaces {
			wsName, _ := ws["name"].(string)
			if !strings.EqualFold(wsName, name) {
				continue
			}
			apiServer, _ := ws["api_server"].(string)
			if apiServer == "" {
				return fmt.Errorf("workspace %q has no api_server — contact support", wsName)
			}
			ctx.Workspace = wsName
			ctx.APIServer = apiServer
			ctx.Idea = ""
			ctx.Person = ""
			ctx.Arc = ""
			ctx.Place = ""
			ctx.Thing = ""
			if err := config.SaveContext(ctx); err != nil {
				return fmt.Errorf("failed to save context: %w", err)
			}
			fmt.Printf("Active workspace: %s\nAPI Endpoint:     %s\n", wsName, apiServer)
			return nil
		}

		return fmt.Errorf("workspace %q not found — run 'yherda workspacelist' to see available workspaces", name)
	},
}
