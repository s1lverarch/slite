package rootfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// iconCross matches main.go's error icon so messages look consistent
// whether they surface from the shell, capsule, rootfs, repo, or config.
const iconCross = "\uf00d" // ✕

// ArchiveType tells the extractor which decompressor to use.
type ArchiveType string

const (
	TarGz  ArchiveType = "tgz"
	TarXz  ArchiveType = "txz"
	TarZst ArchiveType = "tzst"
)

// Entry describes one installable base distro image.
type Entry struct {
	Alias   string // e.g. "saur"
	Name    string // e.g. "Arch Linux"
	URL     string
	Type    ArchiveType
	PkgMgr  string // pacman | apk | apt | dnf | xbps
}

// Registry is the built-in set of supported base distros.
// This mirrors what silverarch-repo publishes; Fetch() below can refresh
// it from the live repo instead of relying only on these hardcoded defaults.
var Registry = map[string]Entry{
	"saur": {
		Alias: "saur", Name: "Arch Linux",
		URL:  "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-x86_64.tar.zst",
		Type: TarZst, PkgMgr: "pacman",
	},
	"salp": {
		Alias: "salp", Name: "Alpine Linux",
		URL:  "https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-3.20.3-x86_64.tar.gz",
		Type: TarGz, PkgMgr: "apk",
	},
	"sdur": {
		Alias: "sdur", Name: "Debian",
		URL:  "https://github.com/debuerreotype/docker-debian-artifacts/raw/dist-amd64/bookworm/rootfs.tar.xz",
		Type: TarXz, PkgMgr: "apt",
	},
	"subu": {
		Alias: "subu", Name: "Ubuntu",
		URL:  "https://cloud-images.ubuntu.com/minimal/releases/noble/release/ubuntu-24.04-minimal-cloudimg-amd64-root.tar.xz",
		Type: TarXz, PkgMgr: "apt",
	},
	"svoid": {
		Alias: "svoid", Name: "Void Linux",
		URL:  "https://repo-default.voidlinux.org/live/current/void-x86_64-ROOTFS-latest.tar.xz",
		Type: TarXz, PkgMgr: "xbps",
	},
	"sfur": {
		Alias: "sfur", Name: "Fedora",
		URL:  "https://dl.fedoraproject.org/pub/fedora/linux/releases/40/Container/x86_64/images/Fedora-Container-Base-40.tar.xz",
		Type: TarXz, PkgMgr: "dnf",
	},
}

// Lookup finds a distro entry by its alias, case-insensitively.
func Lookup(alias string) (Entry, bool) {
	e, ok := Registry[alias]
	return e, ok
}

// ProgressFunc is called repeatedly during download with real byte counts.
// total is 0 if the server didn't send Content-Length (progress bar should
// fall back to a spinner in that case — never fake the percentage).
type ProgressFunc func(downloaded, total int64)

// Fetch downloads (or reuses a cached copy of) a distro's rootfs archive
// and returns the path to the local file. Reports real progress via cb.
func Fetch(cacheDir string, e Entry, cb ProgressFunc) (string, error) {
	dest := filepath.Join(cacheDir, fmt.Sprintf("%s.%s", e.Alias, e.Type))
	if _, err := os.Stat(dest); err == nil {
		return dest, nil // already cached
	}

	tmp := dest + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(e.URL)
	if err != nil {
		os.Remove(tmp)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Remove(tmp)
		return "", fmt.Errorf(iconCross+" download failed: %s (status %d)", e.URL, resp.StatusCode)
	}

	total := resp.ContentLength // -1 if unknown; we normalize to 0 for the caller
	if total < 0 {
		total = 0
	}

	var written int64
	buf := make([]byte, 256*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				os.Remove(tmp)
				return "", werr
			}
			written += int64(n)
			if cb != nil {
				cb(written, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			os.Remove(tmp)
			return "", rerr
		}
	}
	out.Close()

	if err := os.Rename(tmp, dest); err != nil {
		return "", err
	}
	return dest, nil
}
