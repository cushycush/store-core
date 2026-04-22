package hooks

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cushycush/store-core/platform"
)

func TestRunGlobalMissingHook(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := RunGlobal(root, "pre-store", "link"); err != nil {
		t.Fatalf("RunGlobal() error = %v", err)
	}
	assertNotExists(t, filepath.Join(root, "hook-env.txt"))
}

func TestEnvIncludesStoreAndPlatformVars(t *testing.T) {
	t.Parallel()

	root := "/tmp/fake-root"
	env := Env(root, "link", "STORE_EXTRA=x")

	mustContain := []string{
		"STORE_ROOT=" + root,
		"STORE_ACTION=link",
		"STORE_EXTRA=x",
	}
	for _, want := range mustContain {
		if !slices.Contains(env, want) {
			t.Errorf("env missing %q", want)
		}
	}
	for _, want := range platform.Detect().EnvVars() {
		if !slices.Contains(env, want) {
			t.Errorf("env missing platform var %q", want)
		}
	}
}

func writeScript(t *testing.T, path, contents string, mode os.FileMode) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("Chmod(%q) error = %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if !strings.Contains(string(data), want) {
		t.Fatalf("file %q does not contain %q\ncontents:\n%s", path, want, data)
	}
}

func assertTrimmedFileEquals(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if got := strings.TrimSpace(string(data)); got != want {
		t.Fatalf("file %q = %q, want %q", path, got, want)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %q to not exist, got err = %v", path, err)
	}
}

func assertPlatformEnvVars(t *testing.T, path string) {
	t.Helper()

	for _, envVar := range platform.Detect().EnvVars() {
		assertFileContains(t, path, envVar)
	}
}
