package cli

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/store"
)

var exportCSV bool

type exportRow struct {
	Spec   string `json:"spec"`
	Status string `json:"status"`
	Type   string `json:"type"` // spec|req|ac|q|link
	ID     string `json:"id"`
	Text   string `json:"text"`
	Done   bool   `json:"done,omitempty"`
	Rel    string `json:"rel,omitempty"`
	Target string `json:"target,omitempty"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export every spec and item as flat rows (JSON or CSV)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		all, err := store.List(root, flagDir)
		if err != nil {
			return err
		}
		var rows []exportRow
		for _, s := range all {
			ref := s.FM.ID + "-" + s.FM.Slug
			rows = append(rows, exportRow{Spec: ref, Status: s.FM.Status, Type: "spec", ID: s.FM.ID, Text: s.FM.Title})
			for _, r := range s.Reqs {
				rows = append(rows, exportRow{Spec: ref, Status: s.FM.Status, Type: "req", ID: fmt.Sprintf("REQ-%d", r.Num), Text: r.Text})
			}
			for _, c := range s.Criteria {
				rows = append(rows, exportRow{Spec: ref, Status: s.FM.Status, Type: "ac", ID: fmt.Sprintf("AC-%d", c.Num), Text: c.Text, Done: c.Done})
			}
			for _, q := range s.Questions {
				state := "open"
				if q.Resolved {
					state = "resolved"
				} else if q.Blocking {
					state = "blocking"
				}
				rows = append(rows, exportRow{Spec: ref, Status: s.FM.Status, Type: "q", ID: fmt.Sprintf("Q-%d", q.Num), Text: q.Text + " [" + state + "]"})
			}
			for _, l := range s.Links {
				rows = append(rows, exportRow{Spec: ref, Status: s.FM.Status, Type: "link", ID: fmt.Sprintf("L-%d", l.Num), Rel: l.Rel, Target: l.Target})
			}
		}
		if exportCSV {
			w := csv.NewWriter(cmd.OutOrStdout())
			w.Write([]string{"spec", "status", "type", "id", "text", "done", "rel", "target"})
			for _, r := range rows {
				w.Write([]string{r.Spec, r.Status, r.Type, r.ID, r.Text, strconv.FormatBool(r.Done), r.Rel, r.Target})
			}
			w.Flush()
			return w.Error()
		}
		return output.JSON(cmd.OutOrStdout(), rows)
	},
}

func init() {
	exportCmd.Flags().BoolVar(&exportCSV, "csv", false, "CSV instead of JSON")
	rootCmd.AddCommand(exportCmd)
}
