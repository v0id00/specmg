package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
)

var qCmd = &cobra.Command{
	Use:   "q",
	Short: "Manage open questions",
}

func init() { rootCmd.AddCommand(qCmd) }

// --- q add ---

var qAddCmd = &cobra.Command{
	Use:   "add <ref> <text> [--blocking]",
	Short: "Add an open question",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		num := spec.NextQNum(s.Questions)
		s.Questions = append(s.Questions, spec.Question{Num: num, Text: args[1], Blocking: flagQBlocking})
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		if flagDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "would added Q-%d to %s\n", num, s.FM.ID+"-"+s.FM.Slug)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "added Q-%d to %s\n", num, s.FM.ID+"-"+s.FM.Slug)
		}
		return nil
	},
}

// --- q resolve ---

var qResolveCmd = &cobra.Command{
	Use:   "resolve <ref> <Q-N>",
	Short: "Resolve a question",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		typ, n, ok := spec.ParseItemID(args[1])
		if !ok || typ != "Q" {
			return fmt.Errorf("bad item id %q (want Q-N)", args[1])
		}
		for i, q := range s.Questions {
			if q.Num == n {
				s.Questions[i].Resolved = true
				s.Questions[i].Blocking = false
				if err := saveOrDry(root, s); err != nil {
					return err
				}
				if flagDryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would resolved Q-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "resolved Q-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				}
				return nil
			}
		}
		return fmt.Errorf("Q-%d not found in %s", n, s.FM.ID+"-"+s.FM.Slug)
	},
}

// --- q list ---

var qListCmd = &cobra.Command{
	Use:   "list <ref>",
	Short: "List open questions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), s.Questions)
		}
		rows := make([][]string, len(s.Questions))
		for i, q := range s.Questions {
			var state string
			switch {
			case q.Resolved:
				state = "resolved"
			case q.Blocking:
				state = "blocking"
			default:
				state = "open"
			}
			rows[i] = []string{fmt.Sprintf("Q-%d", q.Num), state, q.Text}
		}
		return output.Table(cmd.OutOrStdout(), []string{"ID", "STATE", "TEXT"}, rows)
	},
}

// Flag var for q add
var flagQBlocking bool

func init() {
	qAddCmd.Flags().BoolVar(&flagQBlocking, "blocking", false, "mark as blocking")
	qCmd.AddCommand(qAddCmd, qResolveCmd, qListCmd)
}
