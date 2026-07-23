package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
	"github.com/yherda-opensource/yherda-cmd/internal/structural"
)

var (
	ideasImportFormat string
	ideasImportSource string
	ideasImportIdeaID string
	ideasImportDryRun bool
)

var ideasImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a structural format into an idea's entity graph",
	Long: `Imports entities (identities, arcs, beats, places, things, and documents)
from a structural format such as a Scrivener project into an existing Yherda idea.

This is distinct from 'yherda expression import', which would import rendered
content formats. This command operates on the entity graph.`,
	Example: `  yherda ideas import --format scriv --source ./MyNovel.scriv
  yherda ideas import --format scriv --source ./MyNovel.scriv --idea 42
  yherda ideas import --format scriv --source ./MyNovel.scriv --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ideasImportFormat == "" {
			return fmt.Errorf("--format is required (available: %s)", strings.Join(structural.ImportFormats(), ", "))
		}
		driver, ok := structural.GetImporter(ideasImportFormat)
		if !ok {
			return fmt.Errorf("unknown format %q (available: %s)", ideasImportFormat, strings.Join(structural.ImportFormats(), ", "))
		}

		if ideasImportSource == "" {
			return fmt.Errorf("--source is required")
		}

		ideaID := ideasImportIdeaID
		if ideaID == "" && !ideasImportDryRun {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" && !ideasImportDryRun {
			return fmt.Errorf("no active idea — use --idea <id> or run 'yherda ideas use <id>'")
		}

		return structural.Import(mustClient(), driver, ideaID, ideasImportSource, ideasImportDryRun)
	},
}

func init() {
	ideasImportCmd.Flags().StringVar(&ideasImportFormat, "format", "", "Import format (required): "+strings.Join(structural.ImportFormats(), ", "))
	ideasImportCmd.Flags().StringVar(&ideasImportSource, "source", "", "Path to the source file or directory (required)")
	ideasImportCmd.Flags().StringVar(&ideasImportIdeaID, "idea", "", "Idea ID (overrides active context)")
	ideasImportCmd.Flags().BoolVar(&ideasImportDryRun, "dry-run", false, "Show what would be created without writing to the API")
	// ideas import is disabled — being replaced. See YOS-80.
	// ideasCmd.AddCommand(ideasImportCmd)
}
