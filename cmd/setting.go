package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var settingCmd = &cobra.Command{
	Use:   "setting",
	Short: "Manage settings",
}

var settingPlaceID string

var settingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List settings for a place",
	RunE: func(cmd *cobra.Command, args []string) error {
		placeID := settingPlaceID
		if placeID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			placeID = cfg.ActivePlace
		}
		if placeID == "" {
			fmt.Println("No active place — showing places instead. Run 'yherda place use <id>' to select one.")
			return placeListCmd.RunE(cmd, args)
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/place/"+placeID+"/settings/", &result); err != nil {
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
		printContext()
		return nil
	},
}

func init() {
	settingListCmd.Flags().StringVar(&settingPlaceID, "place", "", "Place ID (overrides active context)")
	settingCmd.AddCommand(settingListCmd)
}
