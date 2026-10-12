package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var (
	flagOwner string
	flagTag   []string
)

var newCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new spec",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := store.NextID(flagRoot, flagDir)
		if err != nil {
			return err
		}
		slug := store.Slugify(args[0])
		if slug == "" {
			return fmt.Errorf("empty slug (provide a meaningful title)")
		}

		now := time.Now().UTC().Format("2006-01-02")
		s := &spec.Spec{
			FM: spec.Frontmatter{
				ID:      id,
				Slug:    slug,
				Title:   args[0],
				Status:  spec.StatusDraft,
				Owner:   flagOwner,
				Created: now,
				Updated: now,
				Tags:    flagTag,
			},
			Purpose:     "TODO",
			DecisionLog: "| Date | Decision | Rationale |\n|------|----------|-----------|",
		}

		filename := s.Filename()
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would create specs/%s\n", filename)
			return nil
		}
		if err := store.Save(flagRoot, flagDir, s); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created specs/%s\n", filename)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&flagOwner, "owner", "", "spec owner")
	newCmd.Flags().StringSliceVar(&flagTag, "tag", nil, "tag (repeatable)")
	rootCmd.AddCommand(newCmd)
}
