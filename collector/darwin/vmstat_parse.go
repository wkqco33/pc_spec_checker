package darwin

import (
	"strconv"
	"strings"
)

// parseVmStatLine extracts the numeric value from a vm_stat output line.
// Example: "Pages free:                    12345." -> 12345
func parseVmStatLine(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	numStr := strings.TrimSuffix(parts[len(parts)-1], ".")
	num, err := strconv.ParseUint(numStr, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
