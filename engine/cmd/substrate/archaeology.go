package main

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/archaeology"
	"github.com/spf13/cobra"
)

func ParseCustomDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)

	// Handle special format "2-years", "1-months", "15-days", etc.
	parts := strings.Split(s, "-")
	if len(parts) == 2 {
		val, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid number before hyphen: %w", err)
		}

		unit := strings.ToLower(parts[1])
		switch unit {
		case "year", "years":
			// Approximate year as 365 days
			return time.Duration(val) * 24 * 365 * time.Hour, nil
		case "month", "months":
			// Approximate month as 30 days
			return time.Duration(val) * 24 * 30 * time.Hour, nil
		case "week", "weeks":
			return time.Duration(val) * 24 * 7 * time.Hour, nil
		case "day", "days":
			return time.Duration(val) * 24 * time.Hour, nil
		}
	}

	// Fallback to standard time.ParseDuration
	return time.ParseDuration(s)
}

var archaeologyCmd = &cobra.Command{
	Use:   "archaeology",
	Short: "Retroactive Dependency Archaeology",
	RunE: func(cmd *cobra.Command, args []string) error {
		sinceStr, _ := cmd.Flags().GetString("since")
		repoPath, _ := cmd.Flags().GetString("repo")
		specPath, _ := cmd.Flags().GetString("spec")
		viper.BindPFlags(cmd.Flags())

		since, err := ParseCustomDuration(sinceStr)
		if err != nil {
			return fmt.Errorf("invalid since format: %w", err)
		}

		fmt.Printf("Running archaeology scan for the past %s...\n", since)

		err = archaeology.Run(repoPath, time.Now().Add(-since), specPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	archaeologyCmd.Flags().String("since", "2-years", "Time duration to look back (e.g., 2-years, 30-days, or 8760h)")
	archaeologyCmd.Flags().String("repo", ".", "Path to the git repository")
	archaeologyCmd.Flags().String("spec", "openapi.yaml", "Path to the spec file within the repo")
	viper.BindPFlags(archaeologyCmd.Flags())
}
