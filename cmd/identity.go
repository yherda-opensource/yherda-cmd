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
	printContext()
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

var identityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new identity for a person",
	RunE: func(cmd *cobra.Command, args []string) error {
		personID, _ := cmd.Flags().GetString("person")
		if personID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			personID = cfg.ActivePerson
		}
		if personID == "" {
			return fmt.Errorf("--person is required (or set active person with 'yherda person use <id>')")
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/role/"+personID+"/identities/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	identityListCmd.Flags().StringVar(&identityPersonID, "person", "", "Person (role) ID (overrides active context)")
	identityCreateCmd.Flags().String("person", "", "Person (role) ID (overrides active context)")
	identityCreateCmd.Flags().String("name", "", "Name of the identity (required)")
	identityCmd.AddCommand(identityListCmd, identityCreateCmd)
}
