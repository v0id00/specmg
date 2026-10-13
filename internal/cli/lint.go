package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/v0id00/specmg/internal/output"
	"github.com/v0id00/specmg/internal/spec"
)

type lintProblem struct {
	File    string `json:"file"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type lintFail struct {
	count int
}

func (l lintFail) Error() string { return fmt.Sprintf("lint: %d problem(s)", l.count) }

var strict bool

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint all specs for consistency issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot()
		if err != nil {
			return err
		}

		all, parseErrs := loadAllSpecs(root)

		var problems []lintProblem
		idSet := make(map[string]bool)
		for _, s := range all {
			key := s.FM.ID + "-" + s.FM.Slug
			idSet[key] = true
		}

		for _, s := range all {
			fname := s.Filename()
			level := "error"

			// (a) approved with unresolved blocking question
			for _, q := range s.Questions {
				if q.Blocking && !q.Resolved && s.FM.Status == "approved" {
					problems = append(problems, lintProblem{
						File:    fname,
						Level:   level,
						Message: fmt.Sprintf("approved spec has unresolved blocking Q-%d", q.Num),
					})
				}
			}

			// (b) superseded without superseded_by
			if s.FM.Status == "superseded" && s.FM.SupersededBy == "" {
				msg := "superseded spec missing superseded_by"
				if strict {
					level = "error"
				} else {
					level = "warning"
				}
				problems = append(problems, lintProblem{
					File:    fname,
					Level:   level,
					Message: msg,
				})
			}

			// (c) dangling links
			for _, lk := range s.Links {
				targetKey := lk.Target
				if !idSet[targetKey] {
					problems = append(problems, lintProblem{
						File:    fname,
						Level:   "error",
						Message: fmt.Sprintf("dangling link L-%d -> %s", lk.Num, targetKey),
					})
				}
			}

			// (d) duplicate item numbers within a type
			dupes := findDupsReq(s.Reqs)
			for _, n := range dupes {
				problems = append(problems, lintProblem{
					File:    fname,
					Level:   "error",
					Message: fmt.Sprintf("duplicate requirement number %d", n),
				})
			}
			dupes = findDupsCrit(s.Criteria)
			for _, n := range dupes {
				problems = append(problems, lintProblem{
					File:    fname,
					Level:   "error",
					Message: fmt.Sprintf("duplicate acceptance criterion number %d", n),
				})
			}
			dupes = findDupsQ(s.Questions)
			for _, n := range dupes {
				problems = append(problems, lintProblem{
					File:    fname,
					Level:   "error",
					Message: fmt.Sprintf("duplicate question number %d", n),
				})
			}
			dupesL := findDupsL(s.Links)
			for _, n := range dupesL {
				problems = append(problems, lintProblem{
					File:    fname,
					Level:   "error",
					Message: fmt.Sprintf("duplicate link number %d", n),
				})
			}
		}

		// Add parse errors
		for _, pe := range parseErrs {
			problems = append(problems, lintProblem{
				File:    pe.file,
				Level:   "error",
				Message: pe.err.Error(),
			})
		}

		// Sort for deterministic output
		sort.Slice(problems, func(i, j int) bool {
			return problems[i].File < problems[j].File
		})

		errCount := 0
		if flagJSON {
			if len(problems) > 0 {
				return output.JSON(cmd.OutOrStdout(), problems)
			}
			return output.JSON(cmd.OutOrStdout(), []lintProblem{})
		}

		for _, p := range problems {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: [%s] %s\n", p.File, p.Level, p.Message)
			if p.Level == "error" {
				errCount++
			}
		}
		if len(problems) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "lint: ok")
		}
		if errCount > 0 {
			return lintFail{count: errCount}
		}
		return nil
	},
}

type parseErrInfo struct {
	file string
	err  error
}

func loadAllSpecs(root string) ([]*spec.Spec, []parseErrInfo) {
	dir := flagDir
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		return nil, []parseErrInfo{{err: err}}
	}
	var specs []*spec.Spec
	var errs []parseErrInfo
	for _, e := range entries {
		if e.IsDir() || !spec.IsSpecFilename(e.Name()) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, dir, e.Name()))
		if err != nil {
			errs = append(errs, parseErrInfo{file: e.Name(), err: err})
			continue
		}
		s, err := spec.Parse(data)
		if err != nil {
			errs = append(errs, parseErrInfo{file: e.Name(), err: err})
			continue
		}
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].FM.ID < specs[j].FM.ID
	})
	return specs, errs
}

func findDupsReq(reqs []spec.Requirement) []int {
	return findDupsNums(func(r spec.Requirement) int { return r.Num }, reqs)
}

func findDupsCrit(criteria []spec.Criterion) []int {
	return findDupsNums(func(c spec.Criterion) int { return c.Num }, criteria)
}

func findDupsQ(questions []spec.Question) []int {
	return findDupsNums(func(q spec.Question) int { return q.Num }, questions)
}

func findDupsL(links []spec.Link) []int {
	return findDupsNums(func(l spec.Link) int { return l.Num }, links)
}

type numExtract[T any] func(T) int

func findDupsNums[T any](extract numExtract[T], items []T) []int {
	seen := make(map[int]int)
	var dups []int
	for _, item := range items {
		n := extract(item)
		if count, ok := seen[n]; ok && count >= 1 {
			continue
		}
		seen[n]++
		if seen[n] == 2 {
			dups = append(dups, n)
		}
	}
	return dups
}

func init() {
	lintCmd.Flags().BoolVarP(&strict, "strict", "s", false, "treat warnings as errors")
	rootCmd.AddCommand(lintCmd)
}
