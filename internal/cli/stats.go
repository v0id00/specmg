package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
	"github.com/v0id00/specmg/internal/store"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Aggregate statistics over all specs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}

		specs, err := store.List(root, flagDir)
		if err != nil {
			return err
		}

		out := aggregate(specs)

		if flagJSON {
			return output.JSON(cmd.OutOrStdout(), out)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "specs: %d   requirements: %d   open ACs: %d   open questions: %d (%d blocking)\n",
			out.SpecsTotal, out.ReqsTotal, out.OpenACs, out.OpenQs, out.BlockingQs)
		for _, kv := range sortedMap(out.ByStatus) {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", kv.K, kv.V)
		}
		return nil
	},
}

type statsOut struct {
	SpecsTotal int            `json:"specs_total"`
	ByStatus   map[string]int `json:"by_status"`
	ReqsTotal  int            `json:"requirements_total"`
	OpenACs    int            `json:"open_acceptance_criteria"`
	OpenQs     int            `json:"open_questions"`
	BlockingQs int            `json:"blocking_questions"`
}

func aggregate(specs []*spec.Spec) statsOut {
	var out statsOut
	out.ByStatus = make(map[string]int)
	for _, s := range specs {
		out.SpecsTotal++
		out.ByStatus[s.FM.Status]++
		out.ReqsTotal += len(s.Reqs)
		for _, c := range s.Criteria {
			if !c.Done {
				out.OpenACs++
			}
		}
		for _, q := range s.Questions {
			if !q.Resolved {
				out.OpenQs++
				if q.Blocking {
					out.BlockingQs++
				}
			}
		}
	}
	return out
}

type kv struct {
	K string
	V int
}

func sortedMap(m map[string]int) []kv {
	var pairs []kv
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].V > pairs[j].V })
	return pairs
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
