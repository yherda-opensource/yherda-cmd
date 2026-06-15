package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace [slug]",
	Short: "Show or set the active workspace",
	Long: `Show the active workspace (no argument) or set it to the given slug.

The active workspace is stored in ~/.yherdacmd/config.json and used by all
subsequent commands. All API calls route to {slug}.a.yherda.com.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(args) == 0 {
			if cfg.ActiveWorkspace == "" {
				fmt.Println("No active workspace set. Run: yherda workspace <slug>")
			} else {
				fmt.Printf("Active workspace: %s\n", cfg.ActiveWorkspace)
			}
			return nil
		}

		cfg.ActiveWorkspace = args[0]
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Active workspace set to: %s\n", args[0])
		return nil
	},
}
