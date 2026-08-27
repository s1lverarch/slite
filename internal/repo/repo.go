package repo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/s1lverarch/slite/internal/rootfs"
)

// ManifestURL points at a JSON file hosted alongside silverarch-repo that
// lists available base-distro rootfs images. This lets Slite pick up new
// distros without a client update — just publish a new manifest entry.
//
// Expected format:
// [
//   {"alias":"saur","name":"Arch Linux","url":"...","type":"tzst","pkgmgr":"pacman"},
//   ...
// ]
const ManifestURL = "https://raw.githubusercontent.com/s1lverarch/silverarch-repo/main/slite/manifest.json"

type manifestEntry struct {
	Alias  string `json:"alias"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Type   string `json:"type"`
	PkgMgr string `json:"pkgmgr"`
}

// Sync fetches the latest manifest from silverarch-repo, caches it locally,
// and merges it into rootfs.Registry (manifest entries override built-in
// defaults with the same alias). Falls back silently to the cached copy —
// or the built-in defaults — if the network is unavailable.
func Sync(repoCacheDir string) error {
	cachePath := filepath.Join(repoCacheDir, "manifest.json")

	data, err := download(ManifestURL)
	if err != nil {
		// offline fallback: use whatever was cached from a previous run
		cached, rerr := os.ReadFile(cachePath)
		if rerr != nil {
			return fmt.Errorf("no network and no cached manifest: %w", err)
		}
		data = cached
	} else {
		_ = os.WriteFile(cachePath, data, 0o644)
	}

	var entries []manifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("parsing manifest: %w", err)
	}

	for _, e := range entries {
		rootfs.Registry[e.Alias] = rootfs.Entry{
			Alias:  e.Alias,
			Name:   e.Name,
			URL:    e.URL,
			Type:   rootfs.ArchiveType(e.Type),
			PkgMgr: e.PkgMgr,
		}
	}
	return nil
}

func download(url string) ([]byte, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest fetch failed: status %d", resp.StatusCode)
	}
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf, nil
}
