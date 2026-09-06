package repo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/s1lverarch/slite/internal/rootfs"
)

// iconCross matches main.go's error icon so messages look consistent
// whether they surface from the shell, capsule, rootfs, repo, or config.
const iconCross = "\uf00d" // ✕

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
			return fmt.Errorf(iconCross+" no network and no cached manifest: %w", err)
		}
		data = cached
	} else {
		_ = os.WriteFile(cachePath, data, 0o644)
	}

	var entries []manifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf(iconCross+" parsing manifest: %w", err)
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
		return nil, fmt.Errorf(iconCross+" manifest fetch failed: status %d", resp.StatusCode)
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

// ReleaseAPIURL is GitHub's "latest release" endpoint for slite itself.
const ReleaseAPIURL = "https://api.github.com/repos/s1lverarch/slite/releases/latest"

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

// LatestVersion checks GitHub for the newest published release tag and
// returns it with any leading "v" stripped (so it compares cleanly against
// the version string baked into the binary at build time).
func LatestVersion() (string, error) {
	data, err := download(ReleaseAPIURL)
	if err != nil {
		return "", err
	}
	var r releaseInfo
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf(iconCross+" parsing release info: %w", err)
	}
	if r.TagName == "" {
		return "", fmt.Errorf(iconCross+" no release tag found")
	}
	return strings.TrimPrefix(r.TagName, "v"), nil
}
