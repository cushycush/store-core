// Package config holds the cross-tool pieces that both store and stock need:
// the .store root layout, when: platform gating, tilde expansion, and root
// discovery.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cushycush/store-core/platform"
)

// ConfigDir is the per-repo directory both store and stock read from.
const ConfigDir = ".store"

// WhenClause gates a config entry on the machine's detected platform. All
// specified fields must match (AND). Each list-valued field accepts a YAML
// scalar or a YAML sequence — within a field, any entry matches (OR). An
// empty field is ignored; a nil WhenClause matches everything.
type WhenClause struct {
	OS            Strings `yaml:"os,omitempty"`
	Arch          Strings `yaml:"arch,omitempty"`
	Distro        Strings `yaml:"distro,omitempty"`
	DistroVersion Strings `yaml:"distro_version,omitempty"`
	Hostname      Strings `yaml:"hostname,omitempty"`
	Shell         Strings `yaml:"shell,omitempty"`
	WSL           *bool   `yaml:"wsl,omitempty"`
}

// Matches returns true if all specified fields match the given platform info.
func (w *WhenClause) Matches(info platform.Info) bool {
	if w == nil {
		return true
	}
	if len(w.OS) > 0 && !w.OS.Contains(info.OS) {
		return false
	}
	if len(w.Arch) > 0 && !w.Arch.Contains(info.Arch) {
		return false
	}
	if len(w.Distro) > 0 && !w.Distro.Contains(info.Distro) {
		return false
	}
	if len(w.DistroVersion) > 0 && !w.DistroVersion.Contains(info.DistroVersion) {
		return false
	}
	if len(w.Hostname) > 0 && !w.Hostname.Contains(info.Hostname) {
		return false
	}
	if len(w.Shell) > 0 && !w.Shell.Contains(info.Shell) {
		return false
	}
	if w.WSL != nil && *w.WSL != info.WSL {
		return false
	}
	return true
}

// FirstMismatch reports the first when: field that doesn't match info and the
// current value for that field. Returns empty strings if the clause matches.
// Useful for building user-facing skip diagnostics.
func (w *WhenClause) FirstMismatch(info platform.Info) (field, want, current string) {
	if w == nil {
		return "", "", ""
	}
	if len(w.OS) > 0 && !w.OS.Contains(info.OS) {
		return "os", w.OS.String(), info.OS
	}
	if len(w.Arch) > 0 && !w.Arch.Contains(info.Arch) {
		return "arch", w.Arch.String(), info.Arch
	}
	if len(w.Distro) > 0 && !w.Distro.Contains(info.Distro) {
		return "distro", w.Distro.String(), info.Distro
	}
	if len(w.DistroVersion) > 0 && !w.DistroVersion.Contains(info.DistroVersion) {
		return "distro_version", w.DistroVersion.String(), info.DistroVersion
	}
	if len(w.Hostname) > 0 && !w.Hostname.Contains(info.Hostname) {
		return "hostname", w.Hostname.String(), info.Hostname
	}
	if len(w.Shell) > 0 && !w.Shell.Contains(info.Shell) {
		return "shell", w.Shell.String(), info.Shell
	}
	if w.WSL != nil && *w.WSL != info.WSL {
		return "wsl", fmt.Sprintf("%t", *w.WSL), fmt.Sprintf("%t", info.WSL)
	}
	return "", "", ""
}

// ExpandHome expands a leading ~ in a path to the user's home directory.
// "~" alone becomes $HOME, "~/x" becomes $HOME/x, and "~user" is returned
// unchanged (no per-user expansion).
func ExpandHome(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return home, nil
	}

	if path[1] == '/' {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

// FindRoot walks up from start looking for a directory containing .store/.
// Returns the directory that contains .store/ (not .store/ itself).
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %q: %w", start, err)
	}

	for {
		if fi, err := os.Stat(filepath.Join(dir, ConfigDir)); err == nil && fi.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no %s directory found", ConfigDir)
}
