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
		if cmd.Flags().Changed("place") {
			useParent(func(cfg *config.Config, id string) { cfg.ActivePlace = id }, placeID)
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

var settingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new setting for a place",
	RunE: func(cmd *cobra.Command, args []string) error {
		placeID, _ := cmd.Flags().GetString("place")
		if placeID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			placeID = cfg.ActivePlace
		}
		if placeID == "" {
			return fmt.Errorf("--place is required (or set active place with 'yherda place use <id>')")
		}
		if cmd.Flags().Changed("place") {
			useParent(func(cfg *config.Config, id string) { cfg.ActivePlace = id }, placeID)
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/place/"+placeID+"/settings/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	settingListCmd.Flags().StringVar(&settingPlaceID, "place", "", "Place ID (overrides active context)")
	settingCreateCmd.Flags().String("place", "", "Place ID (overrides active context)")
	settingCreateCmd.Flags().String("name", "", "Name of the setting (required)")
	settingCmd.AddCommand(settingListCmd, settingCreateCmd)
}
