package ui

import (
	"fmt"
)

// PrintInfo prints a formatted information label and value
func PrintInfo(label string, value string) {
	colors := NewColors()
	fmt.Printf("%s: %s\n", colors.Bold(label), value)
}
