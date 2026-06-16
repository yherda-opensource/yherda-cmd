package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var beatCmd = &cobra.Command{
	Use:   "beat",
	Short: "Manage beats",
}

var beatArcID string

func listBeats(client *api.Client, arcID string) error {
	var result []map[string]any
	if err := client.Get("/arc/"+arcID+"/beats/", &result); err != nil {
		return err
	}
	if jsonOutput {
		printJSON(result)
		return nil
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNUMBER\tDESCRIPTION")
	for _, row := range result {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			strField(row, "id"),
			strField(row, "number"),
			strField(row, "description"),
		)
	}
	w.Flush()
	printContext()
	return nil
}

var beatListCmd = &cobra.Command{
	Use:   "list",
	Short: "List beats for an arc",
	RunE: func(cmd *cobra.Command, args []string) error {
		arcID := beatArcID
		if arcID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			arcID = cfg.ActiveArc
		}
		if arcID == "" {
			fmt.Println("No active arc — showing arcs instead. Run 'yherda arc use <id>' to select one.")
			return arcListCmd.RunE(cmd, args)
		}
		if cmd.Flags().Changed("arc") {
			useParent(func(cfg *config.Config, id string) { cfg.ActiveArc = id }, arcID)
		}
		return listBeats(mustClient(), arcID)
	},
}

var beatCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new beat for an arc",
	RunE: func(cmd *cobra.Command, args []string) error {
		arcID, _ := cmd.Flags().GetString("arc")
		if arcID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			arcID = cfg.ActiveArc
		}
		if arcID == "" {
			return fmt.Errorf("--arc is required (or set active arc with 'yherda arc use <id>')")
		}
		if cmd.Flags().Changed("arc") {
			useParent(func(cfg *config.Config, id string) { cfg.ActiveArc = id }, arcID)
		}
		description, _ := cmd.Flags().GetString("description")
		if description == "" {
			return fmt.Errorf("--description is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/arc/"+arcID+"/beats/", map[string]string{"description": description}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	beatListCmd.Flags().StringVar(&beatArcID, "arc", "", "Arc ID (overrides active context)")
	beatCreateCmd.Flags().String("arc", "", "Arc ID (overrides active context)")
	beatCreateCmd.Flags().String("description", "", "Description of the beat (required)")
	beatCmd.AddCommand(beatListCmd, beatCreateCmd)
}
