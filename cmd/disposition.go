package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var dispositionCmd = &cobra.Command{
	Use:   "disposition",
	Short: "Manage dispositions",
}

var dispositionThingID string

var dispositionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dispositions for a thing",
	RunE: func(cmd *cobra.Command, args []string) error {
		thingID := dispositionThingID
		if thingID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			thingID = cfg.ActiveThing
		}
		if thingID == "" {
			return fmt.Errorf("no active thing — run: yherda thing use <id>")
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/thing/"+thingID+"/dispositions/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				strField(row, "id"),
				strField(row, "name"),
				strField(row, "description"),
			)
		}
		w.Flush()
		return nil
	},
}

func init() {
	dispositionListCmd.Flags().StringVar(&dispositionThingID, "thing", "", "Thing ID (overrides active context)")
	dispositionCmd.AddCommand(dispositionListCmd)
}
