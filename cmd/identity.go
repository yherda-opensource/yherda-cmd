package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Manage identities",
}

var identityPersonID string

func listIdentities(client *api.Client, personID string) error {
	var result []map[string]any
	if err := client.Get("/role/"+personID+"/identities/", &result); err != nil {
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
}

var identityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List identities for a person",
	RunE: func(cmd *cobra.Command, args []string) error {
		personID := identityPersonID
		if personID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			personID = cfg.ActivePerson
		}
		if personID == "" {
			fmt.Println("No active person — showing persons instead. Run 'yherda person use <id>' to select one.")
			return personListCmd.RunE(cmd, args)
		}
		return listIdentities(mustClient(), personID)
	},
}

func init() {
	identityListCmd.Flags().StringVar(&identityPersonID, "person", "", "Person (role) ID (overrides active context)")
	identityCmd.AddCommand(identityListCmd)
}
