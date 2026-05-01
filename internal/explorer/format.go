package explorer

import "fmt"

// FormatBytes converts a byte count to a human-readable string (B, KB, MB, GB, TB, PB).
// Matches the Python version: 2 decimal places, 1024-based units.
func FormatBytes(b float64) string {
	if b == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	for _, unit := range units {
		if b < 1024.0 {
			return fmt.Sprintf("%.2f %s", b, unit)
		}
		b /= 1024.0
	}
	return fmt.Sprintf("%.2f PB", b)
}
