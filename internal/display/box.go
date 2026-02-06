package display

import "fmt"

// Box prints a titled section with visual framing
func Box(title string, content func()) {
	fmt.Printf("┌ %s\n", Bold(title))
	content()
	fmt.Println("└")
}

// Bar renders a proportional bar chart element
func Bar(value, max, width int) string {
	if max <= 0 || value <= 0 {
		return repeat("░", width)
	}
	filled := (value * width) / max
	if filled < 1 && value > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	return Green(repeat("█", filled)) + Dim(repeat("░", width-filled))
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
