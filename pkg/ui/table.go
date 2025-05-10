package ui

import (
	"strings"
)

// TableColumn defines a column in a table
type TableColumn struct {
	Header string
	Width  int
}

// Table represents a simple text-based table
type Table struct {
	Columns   []TableColumn
	Rows      [][]string
	HasHeader bool
}

// NewTable creates a new table with column headers
func NewTable(columns []TableColumn) *Table {
	return &Table{
		Columns:   columns,
		Rows:      [][]string{},
		HasHeader: true,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(values ...string) {
	// Ensure the values match the number of columns
	if len(values) > len(t.Columns) {
		values = values[:len(t.Columns)]
	} else if len(values) < len(t.Columns) {
		// Fill with empty strings
		for i := len(values); i < len(t.Columns); i++ {
			values = append(values, "")
		}
	}
	t.Rows = append(t.Rows, values)
}

// Print prints the table
func (t *Table) Print() {
	if len(t.Columns) == 0 {
		return
	}

	// Print header if it exists
	if t.HasHeader {
		headerString := ""
		for i, col := range t.Columns {
			if i > 0 {
				headerString += " | "
			}
			headerString += padString(col.Header, col.Width)
		}
		Normal("%s\n", headerString)

		// Print separator
		separator := ""
		for i, col := range t.Columns {
			if i > 0 {
				separator += "-+-"
			}
			separator += strings.Repeat("-", col.Width)
		}
		Normal("%s\n", separator)
	}

	// Print rows
	for _, row := range t.Rows {
		rowString := ""
		for i, val := range row {
			if i > 0 {
				rowString += " | "
			}
			rowString += padString(val, t.Columns[i].Width)
		}
		Normal("%s\n", rowString)
	}
}

// padString pads the string to fill the specified width
func padString(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}
