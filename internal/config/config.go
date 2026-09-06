package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Nerd Font icons — match main.go's palette so error messages look
// consistent whether they surface from config, capsule, rootfs, or repo.
const (
	iconCross = "\uf00d" // ✕ error marker
	iconHome  = "\uf015" //  home
)

// Paths holds every on-disk location Slite reads or writes.
type Paths struct {
	Home      string // ~/.slite
	Capsules  string // ~/.slite/capsules
	Cache     string // ~/.slite/cache (downloaded rootfs archives)
	RepoCache string // ~/.slite/repo (mirrored silverarch-repo metadata)
}

// Load resolves Slite's home directory, honoring SLITE_HOME for overrides
// (useful for testing or Termux-style sandboxed HOME setups).
func Load() (*Paths, error) {
	base := os.Getenv("SLITE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf(iconCross+" resolving home directory: %w", err)
		}
		base = filepath.Join(home, ".slite")
	}

	p := &Paths{
		Home:      base,
		Capsules:  filepath.Join(base, "capsules"),
		Cache:     filepath.Join(base, "cache"),
		RepoCache: filepath.Join(base, "repo"),
	}

	for _, dir := range []string{p.Home, p.Capsules, p.Cache, p.RepoCache} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf(iconCross+" "+iconHome+" creating %s: %w", dir, err)
		}
	}
	return p, nil
}
