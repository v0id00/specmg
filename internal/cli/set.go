package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var (
	setStatus string
	setTitle  string
	setOwner  string
	setSlug   string
	setTags   []string
)

var setCmd = &cobra.Command{
	Use:   "set <id|slug>",
	Short: "Update a spec field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}

		oldFilename := s.Filename()

		if setStatus != "" {
			if !spec.ValidStatus(setStatus) {
				return fmt.Errorf("invalid status %q (valid: %v)", setStatus, spec.Statuses)
			}
			s.FM.Status = setStatus
		}
		if setTitle != "" {
			s.FM.Title = setTitle
			if setSlug == "" {
				s.FM.Slug = store.Slugify(setTitle)
			}
		}
		if setOwner != "" {
			s.FM.Owner = setOwner
		}
		if setSlug != "" {
			s.FM.Slug = setSlug
		}
		if len(setTags) > 0 {
			s.FM.Tags = setTags
		}

		newFilename := s.Filename()
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would update %s\n", newFilename)
			return nil
		}
		if err := store.Save(root, flagDir, s); err != nil {
			return err
		}
		if oldFilename != newFilename {
			os.Remove(filepath.Join(root, flagDir, oldFilename))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "updated %s\n", newFilename)
		return nil
	},
}

func init() {
	setCmd.Flags().StringVar(&setStatus, "status", "", "new status")
	setCmd.Flags().StringVar(&setTitle, "title", "", "new title (also regenerates slug unless --slug given)")
	setCmd.Flags().StringVar(&setOwner, "owner", "", "new owner")
	setCmd.Flags().StringVar(&setSlug, "slug", "", "new slug")
	setCmd.Flags().StringSliceVar(&setTags, "tag", nil, "replace tags (repeatable)")
	rootCmd.AddCommand(setCmd)
}
