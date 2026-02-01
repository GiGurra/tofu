package table

// Alignment specifies text alignment within a column
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
)

// Column defines a table column configuration
type Column struct {
	Header   string    // Column header text
	Width    int       // Fixed width (0 = flexible)
	MinWidth int       // Minimum width for flexible columns
	MaxWidth int       // Maximum width (0 = unlimited)
	Weight   float64   // Weight for distributing remaining space (default 1.0)
	Align    Alignment // Left, Right, Center (default Left)
	Truncate bool      // Truncate with ellipsis if content too long
}

// IsFixed returns true if the column has a fixed width
func (c *Column) IsFixed() bool {
	return c.Width > 0
}

// EffectiveWeight returns the weight for flexible width distribution.
// Defaults to 1.0 if not set.
func (c *Column) EffectiveWeight() float64 {
	if c.Weight <= 0 {
		return 1.0
	}
	return c.Weight
}

// CalculateColumnWidths computes actual column widths based on terminal width.
// The algorithm:
// 1. Allocate fixed-width columns first
// 2. Distribute remaining space to flexible columns by weight
// 3. Clamp flexible columns to min/max constraints
// 4. Give any remaining space to the first flexible column that can use it
func CalculateColumnWidths(columns []Column, termWidth int, padding int) []int {
	if len(columns) == 0 {
		return nil
	}

	widths := make([]int, len(columns))

	// Calculate total padding between columns
	totalPadding := padding * (len(columns) - 1)
	availableWidth := max(termWidth-totalPadding, 0)

	// First pass: allocate fixed columns and calculate total weight
	var totalWeight float64
	remainingWidth := availableWidth
	flexibleIndices := []int{}

	for i, col := range columns {
		if col.IsFixed() {
			widths[i] = col.Width
			remainingWidth -= col.Width
		} else {
			totalWeight += col.EffectiveWeight()
			flexibleIndices = append(flexibleIndices, i)
		}
	}

	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Second pass: distribute remaining width to flexible columns by weight
	if totalWeight > 0 && remainingWidth > 0 {
		// Track how much we've actually allocated
		allocated := 0
		for _, i := range flexibleIndices {
			col := columns[i]
			proportion := col.EffectiveWeight() / totalWeight
			width := int(float64(remainingWidth) * proportion)

			// Apply min/max constraints
			if col.MinWidth > 0 && width < col.MinWidth {
				width = col.MinWidth
			}
			if col.MaxWidth > 0 && width > col.MaxWidth {
				width = col.MaxWidth
			}

			widths[i] = width
			allocated += width
		}

		// Give leftover to the last flexible column that can use it
		leftover := remainingWidth - allocated
		if leftover > 0 {
			for j := len(flexibleIndices) - 1; j >= 0; j-- {
				i := flexibleIndices[j]
				col := columns[i]
				canAdd := leftover
				if col.MaxWidth > 0 {
					maxAdd := col.MaxWidth - widths[i]
					if canAdd > maxAdd {
						canAdd = maxAdd
					}
				}
				if canAdd > 0 {
					widths[i] += canAdd
					leftover -= canAdd
				}
				if leftover <= 0 {
					break
				}
			}
		}
	}

	// Ensure minimum widths even if we're over budget
	for i, col := range columns {
		if col.MinWidth > 0 && widths[i] < col.MinWidth {
			widths[i] = col.MinWidth
		}
		// At minimum, every column should have width 1
		if widths[i] < 1 {
			widths[i] = 1
		}
	}

	return widths
}

// FormatCell formats a cell value according to column settings
func FormatCell(value string, col Column, width int) string {
	if col.Truncate && len(value) > width {
		value = TruncateWithEllipsis(value, width)
	}

	switch col.Align {
	case AlignRight:
		return PadLeft(value, width)
	case AlignCenter:
		return PadCenter(value, width)
	default:
		return PadRight(value, width)
	}
}
