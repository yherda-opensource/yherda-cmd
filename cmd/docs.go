package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func readContent(cmd *cobra.Command) (string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file %q: %w", filePath, err)
		}
		return string(data), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("cannot read stdin: %w", err)
	}
	return string(data), nil
}

var docsCmd = &cobra.Command{
	Use:   "doc",
	Short: "Manage idea documents",
	Long:  "An idea document is a freeform note attached to an idea — for research, worldbuilding, or anything outside the structured story model. Documents belong to an idea. (Not to be confused with this CLI's own documentation — see the project README and doc.yherda.com for that.)",
}

var docsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List documents for an idea",
	Example: `  yherda doc list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID, _ := cmd.Flags().GetString("idea")
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			return fmt.Errorf("no active idea — run 'yherda ideas use <id>' or pass --idea")
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/documents/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tTITLE\tUPDATED")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				strField(row, "id"),
				strField(row, "title"),
				strField(row, "updated"))
		}
		w.Flush()
		return nil
	},
}

var docsShowCmd = &cobra.Command{
	Use:     "show <doc-id>",
	Short:   "Show a document's content",
	Example: `  yherda doc show 19`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/ideadocument/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		fmt.Print(strField(result, "body"))
		return nil
	},
}

var docsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new document for an idea",
	Long:  "Creates a new document under an idea. Reads body content from --file, or from stdin if --file is omitted.",
	Example: `  yherda doc create --idea 42 --title "Worldbuilding notes" --file notes.md
  cat notes.md | yherda doc create --idea 42 --title "Worldbuilding notes"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID, _ := cmd.Flags().GetString("idea")
		if ideaID == "" {
			return fmt.Errorf("--idea is required")
		}
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("--title is required")
		}
		body, err := readContent(cmd)
		if err != nil {
			return err
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/storyline/"+ideaID+"/documents/",
			map[string]string{"title": title, "body": body}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var docsUpdateCmd = &cobra.Command{
	Use:   "update <doc-id>",
	Short: "Update a document's content",
	Long:  "Replaces a document's body. Reads from --file, or from stdin if --file is omitted.",
	Example: `  yherda doc update 19 --file revised-notes.md
  cat revised-notes.md | yherda doc update 19`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readContent(cmd)
		if err != nil {
			return err
		}
		client := mustClient()
		var result map[string]any
		if err := client.Patch("/ideadocument/"+args[0]+"/",
			map[string]string{"body": body}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	docsListCmd.Flags().String("idea", "", "Idea ID (overrides active context)")
	docsCreateCmd.Flags().String("idea", "", "Idea ID to create the document under (required)")
	docsCreateCmd.Flags().String("title", "", "Document title (required)")
	docsCreateCmd.Flags().String("file", "", "Path to markdown file (reads stdin if omitted)")
	docsUpdateCmd.Flags().String("file", "", "Path to markdown file (reads stdin if omitted)")
	docsCmd.AddCommand(docsListCmd, docsShowCmd, docsCreateCmd, docsUpdateCmd)
}
