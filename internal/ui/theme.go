package ui

import (
	"fmt"
	"os"
)

// Theme handles ANSI coloring and terminal styling.
type Theme struct {
	Colored bool
}

// NewTheme creates a new theme, honoring the NO_COLOR standard.
func NewTheme(enabled bool) *Theme {
	if !enabled || os.Getenv("NO_COLOR") != "" {
		return &Theme{Colored: false}
	}
	return &Theme{Colored: true}
}

func (t *Theme) colorize(code, s string) string {
	if !t.Colored || s == "" {
		return s
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

func (t *Theme) Bold(s string) string    { return t.colorize("1", s) }
func (t *Theme) Dim(s string) string     { return t.colorize("2", s) }
func (t *Theme) Green(s string) string   { return t.colorize("32", s) }
func (t *Theme) Yellow(s string) string  { return t.colorize("33", s) }
func (t *Theme) Red(s string) string     { return t.colorize("31", s) }
func (t *Theme) Cyan(s string) string    { return t.colorize("36", s) }
func (t *Theme) Blue(s string) string    { return t.colorize("34", s) }
func (t *Theme) Magenta(s string) string { return t.colorize("35", s) }
func (t *Theme) Gray(s string) string    { return t.colorize("90", s) }
func (t *Theme) White(s string) string   { return t.colorize("97", s) }

// UsageColor colors a string based on usage percentage (Green < 60%, Yellow 60-85%, Red > 85%).
func (t *Theme) UsageColor(pct float64, s string) string {
	if !t.Colored {
		return s
	}
	if pct >= 85.0 {
		return t.Red(s)
	}
	if pct >= 60.0 {
		return t.Yellow(s)
	}
	return t.Green(s)
}

// Check returns a checkmark symbol.
func (t *Theme) Check() string {
	if !t.Colored {
		return "✓"
	}
	return t.Green("✓")
}

// Cross returns a cross symbol.
func (t *Theme) Cross() string {
	if !t.Colored {
		return "✗"
	}
	return t.Red("✗")
}

// Bullet returns a bullet symbol.
func (t *Theme) Bullet() string {
	if !t.Colored {
		return "•"
	}
	return t.Cyan("•")
}
