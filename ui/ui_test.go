package ui

import (
	"strings"
	"testing"
)

// resetEnabled installs v for the duration of the test and restores the
// previous value via cleanup.
func resetEnabled(t *testing.T, v bool) {
	t.Helper()
	prev := enabled
	SetEnabled(v)
	t.Cleanup(func() { SetEnabled(prev) })
}

func TestWrapAddsAnsiWhenEnabled(t *testing.T) {
	resetEnabled(t, true)
	got := Bold("x")
	if !strings.HasPrefix(got, "\x1b[1m") || !strings.HasSuffix(got, "\x1b[0m") {
		t.Fatalf("Bold(%q) = %q, want ANSI-wrapped", "x", got)
	}
}

func TestWrapPassthroughWhenDisabled(t *testing.T) {
	resetEnabled(t, false)
	if got := Bold("x"); got != "x" {
		t.Fatalf("Bold(%q) = %q, want %q with styling disabled", "x", got, "x")
	}
}

func TestArrowFallsBackToAsciiWhenDisabled(t *testing.T) {
	resetEnabled(t, false)
	if got := Arrow(); got != "->" {
		t.Fatalf("Arrow() = %q, want %q", got, "->")
	}
}

func TestSemanticHelpersIncludeGlyph(t *testing.T) {
	resetEnabled(t, false) // disable styling so the plain glyph shows through
	tests := []struct {
		name string
		fn   func(string) string
		want string
	}{
		{name: "Success", fn: Success, want: "✓ msg"},
		{name: "Warning", fn: Warning, want: "⚠ msg"},
		{name: "Error", fn: Error, want: "✗ msg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn("msg"); got != tt.want {
				t.Fatalf("%s(%q) = %q, want %q", tt.name, "msg", got, tt.want)
			}
		})
	}
}

func TestDoctorChipsPlainText(t *testing.T) {
	resetEnabled(t, false)
	cases := map[string]string{
		DoctorOK():    "[ok]",
		DoctorWarn():  "[warn]",
		DoctorError(): "[error]",
		DoctorInfo():  "[info]",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("doctor chip = %q, want %q", got, want)
		}
	}
}

func TestPromptFormat(t *testing.T) {
	resetEnabled(t, false)
	got := Prompt("Proceed")
	if got != "Proceed ? " {
		t.Fatalf("Prompt() = %q, want %q", got, "Proceed ? ")
	}
}

func TestEnabledAccessor(t *testing.T) {
	resetEnabled(t, true)
	if !Enabled() {
		t.Fatal("Enabled() = false after SetEnabled(true)")
	}
	SetEnabled(false)
	if Enabled() {
		t.Fatal("Enabled() = true after SetEnabled(false)")
	}
}

func TestBoldVariantsStackCodes(t *testing.T) {
	resetEnabled(t, true)
	// Both "1;31" (bold + red) in one escape is cheaper than nesting two
	// wrappers and makes the output noticeably shorter in doctor reports.
	if got := BoldRed("x"); !strings.Contains(got, "\x1b[1;31m") {
		t.Fatalf("BoldRed emitted %q, expected combined escape", got)
	}
}
