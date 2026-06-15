package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var beatsCmd = &cobra.Command{
	Use:   "beats",
	Short: "Manage beats",
}

var beatsArcID string

var beatsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List beats for an arc",
	RunE: func(cmd *cobra.Command, args []string) error {
		if beatsArcID == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/arcs/"+beatsArcID+"/beats/", &result); err != nil {
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

var beatsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single beat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/beats/"+args[0]+"/", &result); err != nil {
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

var beatsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new beat",
	RunE: func(cmd *cobra.Command, args []string) error {
		arcID, _ := cmd.Flags().GetString("arc")
		name, _ := cmd.Flags().GetString("name")
		if arcID == "" || name == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result map[string]any
		body := map[string]string{"name": name, "arc": arcID}
		if err := client.Post("/beats/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	beatsListCmd.Flags().StringVar(&beatsArcID, "arc", "", "Arc ID (required)")
	beatsCreateCmd.Flags().String("arc", "", "Arc ID (required)")
	beatsCreateCmd.Flags().String("name", "", "Beat name (required)")
	beatsCmd.AddCommand(beatsListCmd, beatsShowCmd, beatsCreateCmd)
}
