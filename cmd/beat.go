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
	if err := client.Get("/arcs/"+arcID+"/beats/", &result); err != nil {
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
		return listBeats(mustClient(), arcID)
	},
}

func init() {
	beatListCmd.Flags().StringVar(&beatArcID, "arc", "", "Arc ID (overrides active context)")
	beatCmd.AddCommand(beatListCmd)
}
