//go:build !windows

package hooks

import (
	"path/filepath"
	"testing"
)

func TestRunGlobalUnix(t *testing.T) {
	t.Parallel()

	const (
		hookName = "pre-store"
		action   = "link"
	)

	tests := []struct {
		name    string
		setup   func(t *testing.T, root string)
		wantErr bool
		check   func(t *testing.T, root string)
	}{
		{
			name: "hook exists and is executable",
			setup: func(t *testing.T, root string) {
				writeScript(t, filepath.Join(root, ".store", "hooks", hookName), "#!/bin/sh\nenv > \"$STORE_ROOT/hook-env.txt\"\npwd > \"$STORE_ROOT/hook-pwd.txt\"\n", 0o755)
			},
			check: func(t *testing.T, root string) {
				assertFileContains(t, filepath.Join(root, "hook-env.txt"), "STORE_ROOT="+root)
				assertFileContains(t, filepath.Join(root, "hook-env.txt"), "STORE_ACTION="+action)
				assertPlatformEnvVars(t, filepath.Join(root, "hook-env.txt"))
				// macOS `pwd` prints the canonicalized path (/private/var/...)
				// while t.TempDir() returns the symlinked /var/... form.
				canonical, err := filepath.EvalSymlinks(root)
				if err != nil {
					t.Fatalf("EvalSymlinks(%q) error = %v", root, err)
				}
				assertTrimmedFileEquals(t, filepath.Join(root, "hook-pwd.txt"), canonical)
			},
		},
		{
			name: "hook exists but is not executable",
			setup: func(t *testing.T, root string) {
				writeScript(t, filepath.Join(root, ".store", "hooks", hookName), "#!/bin/sh\nenv > \"$STORE_ROOT/hook-env.txt\"\n", 0o644)
			},
			check: func(t *testing.T, root string) {
				assertNotExists(t, filepath.Join(root, "hook-env.txt"))
			},
		},
		{
			name: "hook fails",
			setup: func(t *testing.T, root string) {
				writeScript(t, filepath.Join(root, ".store", "hooks", hookName), "#!/bin/sh\nexit 1\n", 0o755)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()

			if tt.setup != nil {
				tt.setup(t, root)
			}

			err := RunGlobal(root, hookName, action)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("RunGlobal() error = %v", err)
			}

			if tt.check != nil {
				tt.check(t, root)
			}
		})
	}
}
