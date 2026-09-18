package ui

import (
	"fmt"
	"os"
	"strconv"
)

// GetTerminalWidth returns the width of the terminal in columns, or defaults to 80.
func GetTerminalWidth() int {
	if w := getTerminalWidth(); w > 0 {
		return w
	}
	if colStr := os.Getenv("COLUMNS"); colStr != "" {
		if c, err := strconv.Atoi(colStr); err == nil && c > 0 {
			return c
		}
	}
	return 80
}

// ClearScreen clears the terminal and moves cursor to home position.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// HideCursor hides the terminal cursor.
func HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor restores the terminal cursor.
func ShowCursor() {
	fmt.Print("\033[?25h")
}
