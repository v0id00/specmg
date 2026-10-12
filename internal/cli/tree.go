package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var treeCmd = &cobra.Command{
	Use:   "tree [<ref>]",
	Short: "Show spec links as a tree; with arg also show backlinks",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		specs, err := store.List(root, flagDir)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			// List every spec and its outgoing links
			for _, s := range specs {
				idSlug := s.FM.ID + "-" + s.FM.Slug
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", idSlug)
				printLinks(cmd.OutOrStdout(), s.Links)
			}
			return nil
		}

		// With a ref: show outgoing AND incoming links
		targetRef := args[0]
		var target *spec.Spec
		for _, s := range specs {
			stem := s.FM.ID + "-" + s.FM.Slug
			if stem == targetRef || s.FM.ID == targetRef || s.FM.Slug == targetRef {
				target = s
				break
			}
		}
		if target == nil {
			return fmt.Errorf("spec %q not found", targetRef)
		}

		idSlug := target.FM.ID + "-" + target.FM.Slug
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", idSlug)
		fmt.Fprintln(cmd.OutOrStdout(), "  outgoing:")
		printLinks(cmd.OutOrStdout(), target.Links)

		// Find incoming (backlinks): other specs whose links target this one
		type incomingLink struct {
			specID string
			l      spec.Link
		}
		var incoming []incomingLink
		for _, s := range specs {
			if s == target {
				continue
			}
			for _, l := range s.Links {
				if l.Target == idSlug || l.Target == target.FM.ID || l.Target == target.FM.Slug {
					incoming = append(incoming, incomingLink{s.FM.ID + "-" + s.FM.Slug, l})
				}
			}
		}
		if len(incoming) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  incoming:")
			for i, in := range incoming {
				last := i == len(incoming)-1
				bar := "├── "
				if last {
					bar = "└── "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s[%s L-%d] %s -> %s\n",
					bar, in.l.Rel, in.l.Num, in.specID, in.l.Target)
			}
		}
		return nil
	},
}

func printLinks(w io.Writer, links []spec.Link) {
	for i, l := range links {
		last := i == len(links)-1
		bar := "├── "
		if last {
			bar = "└── "
		}
		fmt.Fprintf(w, "  %s[L-%d] %s -> %s\n", bar, l.Num, l.Rel, l.Target)
	}
}

func init() {
	rootCmd.AddCommand(treeCmd)
}
