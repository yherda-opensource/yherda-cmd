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
	Long:  "An identity is a belief system a person holds in your story — the lens they see through, which can shift over an arc. Identities belong to a person.",
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
	Long:  "Lists identities for a person. Uses the active person unless --person is passed.",
	Example: `  yherda identity list
  yherda identity list --person 7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		personID := identityPersonID
		if personID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			personID = ctx.Person
		}
		if personID == "" {
			fmt.Println("No active person — showing persons instead. Run 'yherda person use <id>' to select one.")
			return personListCmd.RunE(cmd, args)
		}
		if cmd.Flags().Changed("person") {
			useParent(func(ctx *config.Context, id string) { ctx.Person = id }, personID)
		}
		return listIdentities(mustClient(), personID)
	},
}

var identityCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new identity for a person",
	Long:    "Creates a new identity for a person. Uses the active person unless --person is passed.",
	Example: `  yherda identity create --name "The Skeptic" --person 7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		personID, _ := cmd.Flags().GetString("person")
		if personID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			personID = ctx.Person
		}
		if personID == "" {
			return fmt.Errorf("--person is required (or set active person with 'yherda person use <id>')")
		}
		if cmd.Flags().Changed("person") {
			useParent(func(ctx *config.Context, id string) { ctx.Person = id }, personID)
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
