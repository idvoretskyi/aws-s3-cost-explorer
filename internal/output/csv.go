// Package output provides CSV export helpers.
package output

import (
	"encoding/csv"
	"fmt"
	"os"
)

// WriteCSV writes headers + rows to the given file path, creating or truncating it.
func WriteCSV(path string, headers []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(headers); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
