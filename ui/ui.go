// Package ui provides ANSI color helpers for CLI output shared by store and
// stock. Styling is automatically disabled when stdout is not a terminal, or
// when NO_COLOR is set. FORCE_COLOR overrides both.
//
// This is the CLI styling pipeline. store's TUI has its own lipgloss-based
// palette for the interactive surface — the two systems do not overlap.
package ui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var enabled = detectEnabled()

func detectEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// SetEnabled overrides the auto-detection. Intended for tests; production
// code should rely on the NO_COLOR / FORCE_COLOR environment variables.
func SetEnabled(v bool) { enabled = v }

// Enabled reports whether styling is currently active.
func Enabled() bool { return enabled }

func wrap(code, s string) string {
	if !enabled {
		return s
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

// Typographic weights.
func Bold(s string) string   { return wrap("1", s) }
func Dim(s string) string    { return wrap("2", s) }
func Italic(s string) string { return wrap("3", s) }

// Basic colors. We deliberately stick to the 8-color palette so output looks
// correct in every reasonable terminal, including ones without 256-color
// support.
func Red(s string) string    { return wrap("31", s) }
func Green(s string) string  { return wrap("32", s) }
func Yellow(s string) string { return wrap("33", s) }
func Cyan(s string) string   { return wrap("36", s) }

// Bold variants for headings / emphasis.
func BoldRed(s string) string    { return wrap("1;31", s) }
func BoldGreen(s string) string  { return wrap("1;32", s) }
func BoldYellow(s string) string { return wrap("1;33", s) }
func BoldCyan(s string) string   { return wrap("1;36", s) }

// Semantic helpers used by both tools' long-form output.
func Success(s string) string { return Green("✓ " + s) }
func Warning(s string) string { return Yellow("⚠ " + s) }
func Error(s string) string   { return BoldRed("✗ " + s) }

// Arrow returns a styled → when styling is on, and "->" otherwise so the
// output still lines up in dumb terminals.
func Arrow() string {
	if !enabled {
		return "->"
	}
	return Dim("→")
}

// Doctor chips used by `store doctor` and `stock doctor`. Each has the same
// semantic meaning in either tool.
func DoctorOK() string    { return Green("[ok]") }
func DoctorWarn() string  { return Yellow("[warn]") }
func DoctorError() string { return Red("[error]") }
func DoctorInfo() string  { return Cyan("[info]") }

// Prompt renders an interactive prompt prefix.
func Prompt(question string) string {
	return Bold(question) + " " + Cyan("?") + " "
}
