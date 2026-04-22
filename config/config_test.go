package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cushycush/store-core/platform"
	"gopkg.in/yaml.v3"
)

func TestWhenClauseMatches(t *testing.T) {
	info := platform.Info{
		OS:            "linux",
		Arch:          "amd64",
		Distro:        "ubuntu",
		DistroVersion: "24.04",
		Hostname:      "workstation",
		Shell:         "zsh",
		WSL:           true,
	}

	falseValue := false
	trueValue := true

	tests := []struct {
		name string
		info platform.Info
		when *WhenClause
		want bool
	}{
		{
			name: "nil when clause matches everything",
			info: info,
			when: nil,
			want: true,
		},
		{
			name: "empty when clause matches everything",
			info: info,
			when: &WhenClause{},
			want: true,
		},
		{
			name: "os match",
			info: info,
			when: &WhenClause{OS: Strings{"linux"}},
			want: true,
		},
		{
			name: "os mismatch",
			info: info,
			when: &WhenClause{OS: Strings{"darwin"}},
			want: false,
		},
		{
			name: "os list matches any",
			info: info,
			when: &WhenClause{OS: Strings{"darwin", "linux"}},
			want: true,
		},
		{
			name: "os list mismatches all",
			info: info,
			when: &WhenClause{OS: Strings{"darwin", "windows"}},
			want: false,
		},
		{
			name: "multiple fields all match",
			info: info,
			when: &WhenClause{OS: Strings{"linux"}, Arch: Strings{"amd64"}, Distro: Strings{"ubuntu"}},
			want: true,
		},
		{
			name: "one field mismatches",
			info: info,
			when: &WhenClause{OS: Strings{"linux"}, Arch: Strings{"arm64"}, Distro: Strings{"ubuntu"}},
			want: false,
		},
		{
			name: "wsl true matches",
			info: info,
			when: &WhenClause{WSL: &trueValue},
			want: true,
		},
		{
			name: "wsl false matches",
			info: platform.Info{WSL: false},
			when: &WhenClause{WSL: &falseValue},
			want: true,
		},
		{
			name: "wsl nil matches any",
			info: info,
			when: &WhenClause{OS: Strings{"linux"}, WSL: nil},
			want: true,
		},
		{
			name: "all fields set and matching",
			info: info,
			when: &WhenClause{
				OS:            Strings{"linux"},
				Arch:          Strings{"amd64"},
				Distro:        Strings{"ubuntu"},
				DistroVersion: Strings{"24.04"},
				Hostname:      Strings{"workstation"},
				Shell:         Strings{"zsh"},
				WSL:           &trueValue,
			},
			want: true,
		},
		{
			name: "case insensitive match",
			info: info,
			when: &WhenClause{Distro: Strings{"UBUNTU"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.when.Matches(tt.info); got != tt.want {
				t.Fatalf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhenClauseFirstMismatch(t *testing.T) {
	info := platform.Info{OS: "linux", Arch: "amd64", Distro: "ubuntu", Hostname: "host", Shell: "zsh"}
	trueValue := true

	tests := []struct {
		name                     string
		when                     *WhenClause
		wantField, wantWant, got string
	}{
		{name: "nil", when: nil},
		{name: "match", when: &WhenClause{OS: Strings{"linux"}}},
		{
			name:      "os scalar mismatch",
			when:      &WhenClause{OS: Strings{"darwin"}},
			wantField: "os", wantWant: "darwin", got: "linux",
		},
		{
			name:      "os list mismatch joins values",
			when:      &WhenClause{OS: Strings{"darwin", "windows"}},
			wantField: "os", wantWant: "darwin,windows", got: "linux",
		},
		{
			name:      "wsl mismatch",
			when:      &WhenClause{WSL: &trueValue},
			wantField: "wsl", wantWant: "true", got: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, want, current := tt.when.FirstMismatch(info)
			if field != tt.wantField || want != tt.wantWant || current != tt.got {
				t.Fatalf("FirstMismatch() = (%q, %q, %q), want (%q, %q, %q)",
					field, want, current, tt.wantField, tt.wantWant, tt.got)
			}
		})
	}
}

func TestStringsUnmarshalScalarOrList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Strings
	}{
		{name: "scalar", in: "linux", want: Strings{"linux"}},
		{name: "list", in: "[linux, darwin]", want: Strings{"linux", "darwin"}},
		{name: "empty scalar", in: `""`, want: nil},
		{name: "empty list", in: "[]", want: Strings{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Strings
			if err := yaml.Unmarshal([]byte(tt.in), &got); err != nil {
				t.Fatalf("Unmarshal(%q) error = %v", tt.in, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (got %#v)", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStringsRoundTripMarshalsSingleAsScalar(t *testing.T) {
	tests := []struct {
		name string
		in   Strings
		want string
	}{
		{name: "empty", in: Strings{}, want: "null\n"},
		{name: "single", in: Strings{"linux"}, want: "linux\n"},
		{name: "multi", in: Strings{"linux", "darwin"}, want: "- linux\n- darwin\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := yaml.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(b) != tt.want {
				t.Fatalf("Marshal() = %q, want %q", string(b), tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "empty string", path: "", want: ""},
		{name: "tilde only", path: "~", want: home},
		{name: "tilde slash path", path: "~/foo", want: filepath.Join(home, "foo")},
		{name: "absolute path", path: "/absolute/path", want: "/absolute/path"},
		{name: "tilde username style", path: "~notuser", want: "~notuser"},
		{name: "path without tilde", path: "relative/path", want: "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHome(tt.path)
			if err != nil {
				t.Fatalf("ExpandHome() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExpandHome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindRoot(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (start string, wantRoot string)
		wantErr string
	}{
		{
			name: "finds store in current directory",
			setup: func(t *testing.T) (string, string) {
				root := t.TempDir()
				mustMkdirAll(t, filepath.Join(root, ConfigDir))
				return root, root
			},
		},
		{
			name: "finds store in parent directory",
			setup: func(t *testing.T) (string, string) {
				root := t.TempDir()
				mustMkdirAll(t, filepath.Join(root, ConfigDir))
				child := filepath.Join(root, "a", "b")
				mustMkdirAll(t, child)
				return child, root
			},
		},
		{
			name: "returns error when none found",
			setup: func(t *testing.T) (string, string) {
				return t.TempDir(), ""
			},
			wantErr: "no .store directory found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, wantRoot := tt.setup(t)
			got, err := FindRoot(start)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("FindRoot() error = %v", err)
				}
				// macOS canonicalizes /var/folders → /private/var/folders and
				// Windows returns 8.3 short names. Resolve both sides for a
				// platform-neutral comparison.
				gotResolved, err := filepath.EvalSymlinks(got)
				if err != nil {
					t.Fatalf("EvalSymlinks(%q) error = %v", got, err)
				}
				wantResolved, err := filepath.EvalSymlinks(wantRoot)
				if err != nil {
					t.Fatalf("EvalSymlinks(%q) error = %v", wantRoot, err)
				}
				if gotResolved != wantResolved {
					t.Fatalf("FindRoot() = %q, want %q", got, wantRoot)
				}
				return
			}

			if err == nil {
				t.Fatalf("FindRoot() error = nil, want substring %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("FindRoot() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", path, err)
	}
}
