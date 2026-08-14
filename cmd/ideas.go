package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var ideasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Manage ideas",
	Long:  "An idea is the top-level container for a story — everything else (persons, arcs, beats, places, things, expressions) belongs to one.",
}

func listIdeas(client *api.Client) error {
	var result []map[string]any
	if err := client.Get("/idea/", &result); err != nil {
		return err
	}
	if jsonOutput {
		printJSON(result)
		return nil
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tABSTRACT")
	for _, row := range result {
		fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "name"), strField(row, "abstract"))
	}
	w.Flush()
	printContext()
	return nil
}

var ideasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ideas in the active workspace",
	Example: `  yherda ideas list
  yherda ideas list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listIdeas(mustClient())
	},
}

var ideasShowCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show a single idea",
	Example: `  yherda ideas show 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/idea/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name", "abstract", "idea_type"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

var ideasUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active idea",
	Long:    "Sets the active idea, stored in the .yherda context file. Clears any active person/place/thing/context, since those belong to a specific idea.",
	Example: `  yherda ideas use 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Idea = args[0]
		ctx.Person = ""
		ctx.Place = ""
		ctx.Thing = ""
		ctx.Context = ""
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active idea set to %s\n", args[0])
		return nil
	},
}

func init() {
	ideasCmd.AddCommand(ideasListCmd, ideasShowCmd, ideasUseCmd)
}
