package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/store"
)

var searchCaseSensitive bool

type searchHit struct {
	Ref    string `json:"ref"`
	Status string `json:"status"`
	Item   string `json:"item"` // "title", "REQ-1", "purpose", ...
	Text   string `json:"text"`
}

var searchCmd = &cobra.Command{
	Use:   "search <text>",
	Short: "Search spec titles, purposes and item texts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		all, err := store.List(root, flagDir)
		if err != nil {
			return err
		}
		needle := args[0]
		if !searchCaseSensitive {
			needle = strings.ToLower(needle)
		}
		contains := func(hay string) bool {
			if !searchCaseSensitive {
				hay = strings.ToLower(hay)
			}
			return strings.Contains(hay, needle)
		}
		var hits []searchHit
		icon := func(s string) string { return s }
		if !flagPlain && !flagJSON {
			icon = func(s string) string { return "🔹 " + s }
		}
		for _, s := range all {
			ref := s.FM.ID + "-" + s.FM.Slug
			if contains(s.FM.Title) {
				hits = append(hits, searchHit{ref, s.FM.Status, "title", s.FM.Title})
			}
			if contains(s.Purpose) {
				hits = append(hits, searchHit{ref, s.FM.Status, "purpose", firstLine(s.Purpose)})
			}
			for _, r := range s.Reqs {
				if contains(r.Text) {
					hits = append(hits, searchHit{ref, s.FM.Status, icon(fmt.Sprintf("REQ-%d", r.Num)), r.Text})
				}
			}
			for _, c := range s.Criteria {
				if contains(c.Text) {
					id := fmt.Sprintf("AC-%d", c.Num)
					if !flagPlain && !flagJSON {
						id = output.CriterionIcon(c.Done) + " " + id
					}
					hits = append(hits, searchHit{ref, s.FM.Status, id, c.Text})
				}
			}
			for _, q := range s.Questions {
				if contains(q.Text) {
					id := fmt.Sprintf("Q-%d", q.Num)
					if !flagPlain && !flagJSON {
						id = output.QuestionIcon(q.Blocking, q.Resolved) + " " + id
					}
					hits = append(hits, searchHit{ref, s.FM.Status, id, q.Text})
				}
			}
		}
		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), hits)
		}
		if !flagPlain {
			for i := range hits {
				hits[i].Status = badge(hits[i].Status)
			}
		}
		if len(hits) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no matches")
			return nil
		}
		var rows [][]string
		for _, h := range hits {
			rows = append(rows, []string{h.Ref, badge(h.Status), h.Item, h.Text})
		}
		return output.Table(cmd.OutOrStdout(), []string{"SPEC", "STATUS", "ITEM", "TEXT"}, rows)
	},
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func init() {
	searchCmd.Flags().BoolVar(&searchCaseSensitive, "case", false, "case-sensitive search")
	rootCmd.AddCommand(searchCmd)
}
