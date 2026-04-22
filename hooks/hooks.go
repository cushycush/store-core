// Package hooks runs global scripts under <root>/.store/hooks/ and exposes the
// STORE_* environment contract shared by any per-tool hook runners layered on
// top (for example store's per-store pre/post hooks).
package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cushycush/store-core/platform"
)

// HooksDir is the subdirectory of .store/ where global hook scripts live.
const HooksDir = "hooks"

// RunGlobal executes a global hook script at .store/hooks/<hookName> if it
// exists. Returns nil if the hook does not exist.
//
// On POSIX the script must have the executable bit set and is run directly
// (relying on its shebang). On Windows the executable bit is ignored and
// execution is dispatched by file extension: .ps1 scripts run under
// PowerShell, .cmd/.bat/.exe run directly, and anything else falls back to
// `sh -c <path>` (useful under Git Bash or WSL).
func RunGlobal(root, hookName, action string) error {
	path := filepath.Join(root, ".store", HooksDir, hookName)

	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("hook %q: %w", hookName, err)
	}

	cmd, ok := buildGlobalHookCmd(path, fi)
	if !ok {
		return nil
	}
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = Env(root, action)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q failed: %w", hookName, err)
	}
	return nil
}

// Env returns the STORE_* environment a hook subprocess should see, layered
// onto the current process environment. action is the hook verb (e.g. "link",
// "install"); extra lets callers inject tool-specific variables like
// STORE_NAME or STORE_TARGET.
func Env(root, action string, extra ...string) []string {
	env := append(os.Environ(),
		"STORE_ROOT="+root,
		"STORE_ACTION="+action,
	)
	env = append(env, extra...)
	return append(env, platform.Detect().EnvVars()...)
}

// buildGlobalHookCmd decides how to execute a global hook script file. It
// returns (cmd, true) if the script should run, or (nil, false) if it should
// be silently skipped (e.g. POSIX file without the executable bit).
func buildGlobalHookCmd(path string, fi os.FileInfo) (*exec.Cmd, bool) {
	if runtime.GOOS == "windows" {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ps1":
			return exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", path), true
		case ".cmd", ".bat", ".exe":
			return exec.Command(path), true
		default:
			// Likely a POSIX-style script with a shebang. `sh` is available
			// under Git Bash / WSL / MSYS; if it's missing the command will
			// error at run time and bubble up to the caller.
			return exec.Command("sh", path), true
		}
	}

	if fi.Mode()&0o111 == 0 {
		return nil, false // not executable, skip
	}
	return exec.Command(path), true
}
