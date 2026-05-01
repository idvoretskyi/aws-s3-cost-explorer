// Package output provides table rendering for CLI output.
package output

import (
	"fmt"
	"os"
	"strings"
)

// PrintTable renders rows with given headers in a grid-style table to stdout.
// Uses a simple ASCII grid matching the Python `tabulate` 'grid' format.
func PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := buildSep(widths)
	fmt.Fprintln(os.Stdout, sep)
	fmt.Fprintln(os.Stdout, buildRow(headers, widths))
	fmt.Fprintln(os.Stdout, sep)
	for _, row := range rows {
		fmt.Fprintln(os.Stdout, buildRow(row, widths))
		fmt.Fprintln(os.Stdout, sep)
	}
}

func buildSep(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w+2)
	}
	return "+" + strings.Join(parts, "+") + "+"
}

func buildRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		parts[i] = fmt.Sprintf(" %-*s ", w, cell)
	}
	return "|" + strings.Join(parts, "|") + "|"
}
