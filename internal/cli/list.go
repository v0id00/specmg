package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/store"
)

type listItem struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Status        string   `json:"status"`
	Owner         string   `json:"owner"`
	Tags          []string `json:"tags"`
	Reqs          int      `json:"reqs"`
	ACs           int      `json:"acs"`
	OpenACs       int      `json:"open_acs"`
	Questions     int      `json:"questions"`
	OpenQuestions int      `json:"open_questions"`
}

var (
	listStatus []string
	listOwner  string
	listTag    string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List specs with optional filters",
	RunE: func(cmd *cobra.Command, _ []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}
		all, err := store.List(root, flagDir)
		if err != nil {
			return err
		}

		var items []listItem
		for _, s := range all {
			if len(listStatus) > 0 {
				found := false
				for _, fs := range listStatus {
					if s.FM.Status == fs {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if listOwner != "" && s.FM.Owner != listOwner {
				continue
			}
			if listTag != "" {
				tagFound := false
				for _, t := range s.FM.Tags {
					if t == listTag {
						tagFound = true
						break
					}
				}
				if !tagFound {
					continue
				}
			}

			openAC := 0
			for _, c := range s.Criteria {
				if !c.Done {
					openAC++
				}
			}
			openQ := 0
			for _, q := range s.Questions {
				if !q.Resolved {
					openQ++
				}
			}

			items = append(items, listItem{
				ID:            s.FM.ID,
				Slug:          s.FM.Slug,
				Title:         s.FM.Title,
				Status:        s.FM.Status,
				Owner:         s.FM.Owner,
				Tags:          s.FM.Tags,
				Reqs:          len(s.Reqs),
				ACs:           len(s.Criteria),
				OpenACs:       openAC,
				Questions:     len(s.Questions),
				OpenQuestions: openQ,
			})
		}

		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), items)
		}

		header := []string{"ID", "STATUS", "TITLE", "OWNER", "REQS", "OPEN-AC", "OPEN-Q"}
		var rows [][]string
		for _, it := range items {
			rows = append(rows, []string{
				it.ID, badge(it.Status), it.Title, it.Owner,
				fmt.Sprintf("%d", it.Reqs),
				fmt.Sprintf("%d", it.OpenACs),
				fmt.Sprintf("%d", it.OpenQuestions),
			})
		}
		return output.Table(cmd.OutOrStdout(), header, rows)
	},
}

func init() {
	listCmd.Flags().StringSliceVar(&listStatus, "status", nil, "filter by status (repeatable)")
	listCmd.Flags().StringVar(&listOwner, "owner", "", "filter by owner")
	listCmd.Flags().StringVar(&listTag, "tag", "", "filter by tag")
	rootCmd.AddCommand(listCmd)
}
