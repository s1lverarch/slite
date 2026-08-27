package capsule

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/s1lverarch/slite/internal/rootfs"
)

// Manager operates on capsules rooted at a given capsules directory.
type Manager struct {
	CapsulesDir string
	CacheDir    string
}

func New(capsulesDir, cacheDir string) *Manager {
	return &Manager{CapsulesDir: capsulesDir, CacheDir: cacheDir}
}

func (m *Manager) path(name string) string {
	return filepath.Join(m.CapsulesDir, name)
}

// Exists reports whether a capsule with this name has already been created.
func (m *Manager) Exists(name string) bool {
	_, err := os.Stat(m.path(name))
	return err == nil
}

// Create downloads (if needed) and extracts a base distro rootfs into a new
// capsule directory, then writes a marker file recording which package
// manager that capsule uses.
func (m *Manager) Create(alias, name string) error {
	entry, ok := rootfs.Lookup(alias)
	if !ok {
		return fmt.Errorf("unknown base distro alias %q", alias)
	}
	if m.Exists(name) {
		return fmt.Errorf("capsule %q already exists", name)
	}

	archive, err := rootfs.Fetch(m.CacheDir, entry)
	if err != nil {
		return fmt.Errorf("fetching rootfs: %w", err)
	}

	target := m.path(name)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}

	if err := extract(archive, entry.Type, target); err != nil {
		os.RemoveAll(target)
		return fmt.Errorf("extracting rootfs: %w", err)
	}

	// Arch's bootstrap tarball nests everything under root.x86_64/
	nested := filepath.Join(target, "root.x86_64")
	if info, err := os.Stat(nested); err == nil && info.IsDir() {
		if err := flatten(nested, target); err != nil {
			return fmt.Errorf("flattening arch bootstrap: %w", err)
		}
	}

	// Copy host DNS resolution in so networking works inside the capsule.
	os.MkdirAll(filepath.Join(target, "etc"), 0o755)
	if data, err := os.ReadFile("/etc/resolv.conf"); err == nil {
		os.WriteFile(filepath.Join(target, "etc", "resolv.conf"), data, 0o644)
	}

	return os.WriteFile(filepath.Join(target, ".slite-pkgmgr"), []byte(entry.PkgMgr), 0o644)
}

func extract(archive string, t rootfs.ArchiveType, dest string) error {
	var flag string
	switch t {
	case rootfs.TarGz:
		flag = "-xzf"
	case rootfs.TarXz:
		flag = "-xJf"
	case rootfs.TarZst:
		flag = "--zstd"
	default:
		return fmt.Errorf("unsupported archive type %q", t)
	}

	var cmd *exec.Cmd
	if t == rootfs.TarZst {
		cmd = exec.Command("tar", "--zstd", "-xf", archive, "-C", dest)
	} else {
		cmd = exec.Command("tar", flag, archive, "-C", dest)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = flag
	return cmd.Run()
}

func flatten(nested, target string) error {
	entries, err := os.ReadDir(nested)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.Rename(filepath.Join(nested, e.Name()), filepath.Join(target, e.Name())); err != nil {
			return err
		}
	}
	return os.Remove(nested)
}

// PkgMgr returns the package manager recorded for a capsule at creation time.
func (m *Manager) PkgMgr(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(m.path(name), ".slite-pkgmgr"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Exec runs a command inside a capsule using proot for rootless syscall
// emulation. If cmdArgs is empty it drops into an interactive login shell.
func (m *Manager) Exec(name string, cmdArgs []string) error {
	if !m.Exists(name) {
		return fmt.Errorf("capsule %q not found", name)
	}
	target := m.path(name)

	if len(cmdArgs) == 0 {
		cmdArgs = []string{"bash", "-l"}
	}

	args := []string{
		"-0", // fake root inside the capsule
		"-r", target,
		"-b", "/dev",
		"-b", "/proc",
		"-b", "/sys",
		"-b", filepath.Join(target, "etc", "resolv.conf") + ":/etc/resolv.conf",
		"-w", "/root",
		"--kill-on-exit",
	}
	args = append(args, cmdArgs...)

	prootPath, err := exec.LookPath("proot")
	if err != nil {
		return fmt.Errorf("proot not found in PATH — install it first")
	}

	// Replace this process image entirely (like exec.Cmd but with true exec(2))
	// so signals and the TTY behave exactly as if proot had been run directly.
	env := os.Environ()
	return syscall.Exec(prootPath, append([]string{"proot"}, args...), env)
}

// InstallCmd builds the package-manager-specific install command for a capsule.
func (m *Manager) InstallCmd(pkgMgr, pkg string) ([]string, error) {
	switch pkgMgr {
	case "pacman":
		return []string{"bash", "-lc", "pacman-key --init >/dev/null 2>&1; pacman-key --populate archlinux >/dev/null 2>&1; pacman -Sy --noconfirm " + pkg}, nil
	case "apk":
		return []string{"bash", "-lc", "apk add " + pkg}, nil
	case "apt":
		return []string{"bash", "-lc", "apt-get update && apt-get install -y " + pkg}, nil
	case "dnf":
		return []string{"bash", "-lc", "dnf install -y " + pkg}, nil
	case "xbps":
		return []string{"bash", "-lc", "xbps-install -Sy " + pkg}, nil
	default:
		return nil, fmt.Errorf("no install command known for package manager %q", pkgMgr)
	}
}

// List returns the names of all existing capsules.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.CapsulesDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// Remove deletes a capsule entirely.
func (m *Manager) Remove(name string) error {
	if !m.Exists(name) {
		return fmt.Errorf("capsule %q not found", name)
	}
	return os.RemoveAll(m.path(name))
}
