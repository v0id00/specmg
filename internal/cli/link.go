package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Manage links between specs",
}

func init() { rootCmd.AddCommand(linkCmd) }

// --- link add ---

var linkAddCmd = &cobra.Command{
	Use:   "add <ref> --rel <rel> --to <target>",
	Short: "Add a link to another spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return usageErr{fmt.Errorf("usage: link add <ref> --rel <rel> --to <target>")}
		}
		rel := flagLinkRel
		if rel == "" {
			return usageErr{fmt.Errorf("--rel is required")}
		}
		to := flagLinkTo
		if to == "" {
			return usageErr{fmt.Errorf("--to is required")}
		}
		if !spec.ValidRel(rel) {
			return fmt.Errorf("invalid rel %q (valid: implements, relates, blocks, depends_on, supersedes)", rel)
		}
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		target, err := store.Load(root, flagDir, to)
		if err != nil {
			return fmt.Errorf("target %q not found", to)
		}
		num := spec.NextLNum(s.Links)
		s.Links = append(s.Links, spec.Link{Num: num, Rel: rel, Target: target.FM.ID + "-" + target.FM.Slug})
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would added L-%d to %s\n", num, s.FM.ID+"-"+s.FM.Slug)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "added L-%d to %s\n", num, s.FM.ID+"-"+s.FM.Slug)
		}
		return nil
	},
}

// --- link rm ---

var linkRmCmd = &cobra.Command{
	Use:   "rm <ref> <L-N>",
	Short: "Remove a link",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		typ, n, ok := spec.ParseItemID(args[1])
		if !ok || typ != "L" {
			return fmt.Errorf("bad item id %q (want L-N)", args[1])
		}
		for i, l := range s.Links {
			if l.Num == n {
				s.Links = append(s.Links[:i], s.Links[i+1:]...)
				if err := saveOrDry(root, s); err != nil {
					return err
				}
				if flagDryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would removed L-%d from %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "removed L-%d from %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				}
				return nil
			}
		}
		return fmt.Errorf("L-%d not found in %s", n, s.FM.ID+"-"+s.FM.Slug)
	},
}

// --- link list ---

var linkListCmd = &cobra.Command{
	Use:   "list <ref>",
	Short: "List links",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), s.Links)
		}
		rows := make([][]string, len(s.Links))
		for i, l := range s.Links {
			rows[i] = []string{fmt.Sprintf("L-%d", l.Num), l.Rel, l.Target}
		}
		return output.Table(cmd.OutOrStdout(), []string{"ID", "REL", "TARGET"}, rows)
	},
}

// Flag vars for link add
var (
	flagLinkRel string
	flagLinkTo  string
)

func init() {
	linkAddCmd.Flags().StringVarP(&flagLinkRel, "rel", "r", "", "link relation")
	linkAddCmd.Flags().StringVarP(&flagLinkTo, "to", "t", "", "target spec ref")
	_ = linkAddCmd.MarkFlagRequired("rel")
	_ = linkAddCmd.MarkFlagRequired("to")
	linkCmd.AddCommand(linkAddCmd, linkRmCmd, linkListCmd)
}
