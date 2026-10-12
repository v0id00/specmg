package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
)

var reqCmd = &cobra.Command{
	Use:   "req",
	Short: "Manage requirements",
}

func init() { rootCmd.AddCommand(reqCmd) }

// --- req add ---

var reqAddCmd = &cobra.Command{
	Use:   "add <ref> <text>",
	Short: "Add a requirement",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		num := spec.NextReqNum(s.Reqs)
		s.Reqs = append(s.Reqs, spec.Requirement{Num: num, Text: args[1]})
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		prefix := ""
		if flagDryRun {
			prefix = "would "
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%sadded REQ-%d to %s\n", prefix, num, s.FM.ID+"-"+s.FM.Slug)
		return nil
	},
}

// --- req rm ---

var reqRmCmd = &cobra.Command{
	Use:   "rm <ref> <REQ-N>",
	Short: "Remove a requirement",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		typ, n, ok := spec.ParseItemID(args[1])
		if !ok || typ != "REQ" {
			return fmt.Errorf("bad item id %q (want REQ-N)", args[1])
		}
		for i, r := range s.Reqs {
			if r.Num == n {
				s.Reqs = append(s.Reqs[:i], s.Reqs[i+1:]...)
				if err := saveOrDry(root, s); err != nil {
					return err
				}
				if flagDryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would removed REQ-%d from %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "removed REQ-%d from %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				}
				return nil
			}
		}
		return fmt.Errorf("REQ-%d not found in %s", n, s.FM.ID+"-"+s.FM.Slug)
	},
}

// --- req list ---

var reqListCmd = &cobra.Command{
	Use:   "list <ref>",
	Short: "List requirements",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), s.Reqs)
		}
		rows := make([][]string, len(s.Reqs))
		for i, r := range s.Reqs {
			rows[i] = []string{fmt.Sprintf("REQ-%d", r.Num), r.Text}
		}
		return output.Table(cmd.OutOrStdout(), []string{"ID", "TEXT"}, rows)
	},
}

func init() {
	reqCmd.AddCommand(reqAddCmd, reqRmCmd, reqListCmd)
}
