package benchmark

import (
	"fmt"
	"strings"
	"time"
)

// printResultsTable prints a formatted table of benchmark results.
func printResultsTable(results []Result) {
	if len(results) == 0 {
		return
	}

	// Calculate column widths
	solverWidth := 8 // minimum "Solver"
	for _, r := range results {
		if len(r.SolverName) > solverWidth {
			solverWidth = len(r.SolverName)
		}
	}

	avgTimeWidth := 10  // "Avg Time"
	totalWidth := 10    // "Total"
	solvedWidth := 8    // "Solved"
	optimalWidth := 9   // "Optimal"

	// Print header
	printTableBorder(solverWidth, avgTimeWidth, totalWidth, solvedWidth, optimalWidth, "top")
	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		solverWidth, "Solver",
		avgTimeWidth, "Avg Time",
		totalWidth, "Total",
		solvedWidth, "Solved",
		optimalWidth, "Optimal")
	printTableBorder(solverWidth, avgTimeWidth, totalWidth, solvedWidth, optimalWidth, "mid")

	// Print rows
	for _, r := range results {
		fmt.Printf("│ %-*s │ %*s │ %*s │ %*s │ %*s │\n",
			solverWidth, r.SolverName,
			avgTimeWidth, formatDuration(r.AvgTime),
			totalWidth, formatDuration(r.TotalTime),
			solvedWidth, fmt.Sprintf("%d/%d", r.SolvedCount, r.PuzzleCount),
			optimalWidth, fmt.Sprintf("%d/%d", r.OptimalCount, r.PuzzleCount))
	}

	printTableBorder(solverWidth, avgTimeWidth, totalWidth, solvedWidth, optimalWidth, "bottom")
}

// printTableBorder prints a table border line.
func printTableBorder(solverW, avgW, totalW, solvedW, optimalW int, position string) {
	var left, mid, right string
	horiz := "─"
	switch position {
	case "top":
		left, mid, right = "┌", "┬", "┐"
	case "mid":
		left, mid, right = "├", "┼", "┤"
	case "bottom":
		left, mid, right = "└", "┴", "┘"
	}

	fmt.Printf("%s%s%s%s%s%s%s%s%s%s%s\n",
		left, strings.Repeat(horiz, solverW+2),
		mid, strings.Repeat(horiz, avgW+2),
		mid, strings.Repeat(horiz, totalW+2),
		mid, strings.Repeat(horiz, solvedW+2),
		mid, strings.Repeat(horiz, optimalW+2),
		right)
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
