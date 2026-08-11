// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package archaeology

import (
	"fmt"
	"os"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// IncidentCost represents the estimated cost of an incident based on breaking changes.
type IncidentCost struct {
	TotalCost float64
	Count     int
}

// EstimateCost calculates the cost based on the number of breaking changes.
func EstimateCost(breakingCount int) IncidentCost {
	// A basic simple cost multiplier matching "P13-T01" estimation logic.
	// We'll assign $5,000 per breaking change for estimation.
	return IncidentCost{
		TotalCost: float64(breakingCount) * 5000.0,
		Count:     breakingCount,
	}
}

// Run scans the git repository starting from the given path
func Run(repoPath string, since time.Time, specPath string) error {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	ref, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get head: %w", err)
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	var previousFileContent string
	var previousCommitHash string
	var previousCommitDate time.Time
	var totalCost float64
	var totalBreakingChanges int

	fmt.Println("# Retroactive Dependency Archaeology Report")
	fmt.Println("Analyzing historical breaking changes...")
	fmt.Println()

	err = cIter.ForEach(func(c *object.Commit) error {
		tree, err := c.Tree()
		if err != nil {
			if c.Author.When.Before(since) {
				return storer.ErrStop
			}
			return nil
		}

		file, err := tree.File(specPath)
		if err != nil {
			if c.Author.When.Before(since) {
				return storer.ErrStop
			}
			return nil
		}

		content, err := file.Contents()
		if err != nil {
			if c.Author.When.Before(since) {
				return storer.ErrStop
			}
			return nil
		}

		// If the file contents are identical, no need to diff.
		if previousFileContent == content {
			previousCommitHash = c.Hash.String()
			previousCommitDate = c.Author.When
			if c.Author.When.Before(since) {
				return storer.ErrStop
			}
			return nil
		}

		if previousFileContent != "" {
			// Write temporary files to diff
			baseFile, err := os.CreateTemp("", "base-*.yaml")
			if err != nil {
				if c.Author.When.Before(since) {
					return storer.ErrStop
				}
				return nil
			}
			defer os.Remove(baseFile.Name())
			baseFile.WriteString(content)
			baseFile.Close()

			revFile, err := os.CreateTemp("", "rev-*.yaml")
			if err != nil {
				if c.Author.When.Before(since) {
					return storer.ErrStop
				}
				return nil
			}
			defer os.Remove(revFile.Name())
			revFile.WriteString(previousFileContent)
			revFile.Close()

			rep, err := diff.CompareOpenAPI(baseFile.Name(), revFile.Name(), true, nil)
			if err == nil && rep != nil && rep.Summary.BreakingCount > 0 {
				cost := EstimateCost(rep.Summary.BreakingCount)
				totalCost += cost.TotalCost
				totalBreakingChanges += cost.Count

				fmt.Printf("## Incident at %s\n", previousCommitHash)
				fmt.Printf("- **Date:** %s\n", previousCommitDate.Format(time.RFC3339))
				fmt.Printf("- **Breaking Changes:** %d\n", rep.Summary.BreakingCount)
				fmt.Printf("- **Estimated Incident Cost:** $%.2f\n\n", cost.TotalCost)
			}
		}

		previousFileContent = content
		previousCommitHash = c.Hash.String()
		previousCommitDate = c.Author.When

		if c.Author.When.Before(since) {
			return storer.ErrStop
		}
		return nil
	})

	if err != nil && err != storer.ErrStop {
		return err
	}

	fmt.Println("---")
	fmt.Println("## Summary: Hidden API Debt")
	fmt.Printf("- **Total Breaking Changes:** %d\n", totalBreakingChanges)
	fmt.Printf("- **Total Estimated Cost:** $%.2f\n", totalCost)

	return nil
}
