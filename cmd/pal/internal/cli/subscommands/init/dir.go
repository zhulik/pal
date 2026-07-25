package initcmd

import (
	"fmt"
	"os"
)

// IsEmpty reports whether dir contains no entries.
func IsEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read directory: %w", err)
	}
	return len(entries) == 0, nil
}
