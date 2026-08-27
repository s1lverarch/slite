package config

import (
	"os"
	"path/filepath"
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
			return nil, err
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
			return nil, err
		}
	}
	return p, nil
}
