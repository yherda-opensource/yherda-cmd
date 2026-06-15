package cmd

import (
	"github.com/spf13/cobra"
)

var identitiesCmd = &cobra.Command{
	Use:   "identities",
	Short: "Manage identities",
}

var identitiesIdeaID string

var identitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List identities for an idea",
	RunE: func(cmd *cobra.Command, args []string) error {
		if identitiesIdeaID == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result any
		if err := client.Get("/ideas/"+identitiesIdeaID+"/identities/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var identitiesShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single identity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result any
		if err := client.Get("/identities/"+args[0]+"/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var identitiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new identity",
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID, _ := cmd.Flags().GetString("idea")
		name, _ := cmd.Flags().GetString("name")
		if ideaID == "" || name == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result any
		body := map[string]string{"name": name, "idea": ideaID}
		if err := client.Post("/identities/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	identitiesListCmd.Flags().StringVar(&identitiesIdeaID, "idea", "", "Idea ID (required)")
	identitiesCreateCmd.Flags().String("idea", "", "Idea ID (required)")
	identitiesCreateCmd.Flags().String("name", "", "Identity name (required)")
	identitiesCmd.AddCommand(identitiesListCmd, identitiesShowCmd, identitiesCreateCmd)
}
