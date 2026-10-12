package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
)

var acCmd = &cobra.Command{
	Use:   "ac",
	Short: "Manage acceptance criteria",
}

func init() { rootCmd.AddCommand(acCmd) }

// --- ac add ---

var acAddCmd = &cobra.Command{
	Use:   "add <ref> <text>",
	Short: "Add an acceptance criterion",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		num := spec.NextACNum(s.Criteria)
		s.Criteria = append(s.Criteria, spec.Criterion{Num: num, Text: args[1], Done: false})
		if err := saveOrDry(root, s); err != nil {
			return err
		}
		prefix := ""
		if flagDryRun {
			prefix = "would "
		}
		id := s.FM.ID + "-" + s.FM.Slug
		fmt.Fprintf(cmd.OutOrStdout(), "%sadded AC-%d to %s\n", prefix, num, id)
		return nil
	},
}

// --- ac check ---

var acCheckCmd = &cobra.Command{
	Use:   "check <ref> <AC-N>",
	Short: "Mark an acceptance criterion as done",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		typ, n, ok := spec.ParseItemID(args[1])
		if !ok || typ != "AC" {
			return fmt.Errorf("bad item id %q (want AC-N)", args[1])
		}
		for i, c := range s.Criteria {
			if c.Num == n {
				s.Criteria[i].Done = true
				if err := saveOrDry(root, s); err != nil {
					return err
				}
				if flagDryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would checked AC-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "checked AC-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				}
				return nil
			}
		}
		return fmt.Errorf("AC-%d not found in %s", n, s.FM.ID+"-"+s.FM.Slug)
	},
}

// --- ac uncheck ---

var acUncheckCmd = &cobra.Command{
	Use:   "uncheck <ref> <AC-N>",
	Short: "Mark an acceptance criterion as undone",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		typ, n, ok := spec.ParseItemID(args[1])
		if !ok || typ != "AC" {
			return fmt.Errorf("bad item id %q (want AC-N)", args[1])
		}
		for i, c := range s.Criteria {
			if c.Num == n {
				s.Criteria[i].Done = false
				if err := saveOrDry(root, s); err != nil {
					return err
				}
				if flagDryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would unchecked AC-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "unchecked AC-%d on %s\n", n, s.FM.ID+"-"+s.FM.Slug)
				}
				return nil
			}
		}
		return fmt.Errorf("AC-%d not found in %s", n, s.FM.ID+"-"+s.FM.Slug)
	},
}

// --- ac list ---

var acListCmd = &cobra.Command{
	Use:   "list <ref>",
	Short: "List acceptance criteria",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := mustLoad(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), s.Criteria)
		}
		rows := make([][]string, len(s.Criteria))
		for i, c := range s.Criteria {
			rows[i] = []string{fmt.Sprintf("AC-%d", c.Num), check(c.Done), c.Text}
		}
		return output.Table(cmd.OutOrStdout(), []string{"ID", "DONE", "TEXT"}, rows)
	},
}

func init() {
	acCmd.AddCommand(acAddCmd, acCheckCmd, acUncheckCmd, acListCmd)
}
