package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
	"github.com/yherda-opensource/yherda-cmd/internal/structural"
)

var (
	ideasExportFormat string
	ideasExportIdeaID string
	ideasExportOutput string
)

var ideasExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export an idea's entity graph to a structural format",
	Long: `Exports the entity graph of an idea (identities, arcs, beats, places, things,
and attached documents) to a structural format such as an Obsidian vault.

This is distinct from 'yherda expression export', which exports a segment tree
into a rendered content format (manuscript, screenplay, etc.).`,
	Example: `  yherda ideas export --format obsidian
  yherda ideas export --format obsidian --idea 42 --output ~/vaults/my-idea`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ideasExportFormat == "" {
			return fmt.Errorf("--format is required (available: %s)", strings.Join(structural.Formats(), ", "))
		}
		driver, ok := structural.Get(ideasExportFormat)
		if !ok {
			return fmt.Errorf("unknown format %q (available: %s)", ideasExportFormat, strings.Join(structural.Formats(), ", "))
		}

		ideaID := ideasExportIdeaID
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			return fmt.Errorf("no active idea — use --idea <id> or run 'yherda ideas use <id>'")
		}

		// Build the manifest from what the user asked for.
		// Today we always export everything; a future --entities flag would
		// let the user control which entity types are included.
		m := structural.Manifest{
			Identities: true,
			Arcs:       true,
			Beats:      true,
			Places:     true,
			Things:     true,
			Docs:       true,
		}

		return structural.Export(mustClient(), driver, m, ideaID, ideasExportOutput)
	},
}

func init() {
	ideasExportCmd.Flags().StringVar(&ideasExportFormat, "format", "", "Export format (required): "+strings.Join(structural.Formats(), ", "))
	ideasExportCmd.Flags().StringVar(&ideasExportIdeaID, "idea", "", "Idea ID (overrides active context)")
	ideasExportCmd.Flags().StringVar(&ideasExportOutput, "output", "", "Output path (default: ./obsidian-export/)")
	ideasCmd.AddCommand(ideasExportCmd)
}
