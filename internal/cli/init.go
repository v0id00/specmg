package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the specs/ directory if missing",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			cwd, cerr := os.Getwd()
			if cerr != nil {
				return cerr
			}
			root = cwd // fresh repo: init where we stand
		}
		dir := filepath.Join(root, flagDir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "initialized %s\n", dir)
		return nil
	},
}

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "specmg "+version)
	},
}

func init() { rootCmd.AddCommand(initCmd, versionCmd) }
