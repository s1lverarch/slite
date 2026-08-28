package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/s1lverarch/slite/internal/capsule"
	"github.com/s1lverarch/slite/internal/config"
	"github.com/s1lverarch/slite/internal/repo"
	"github.com/s1lverarch/slite/internal/rootfs"
)

const banner = `
  ________  ___       ___  _________  _______
|\   ____\|\  \     |\  \|\___   ___\\  ___ \
\ \  \___|\ \  \    \ \  \|___ \  \_\ \   __/|
 \ \_____  \ \  \    \ \  \   \ \  \ \ \  \_|/__
  \|____|\  \ \  \____\ \  \   \ \  \ \ \  \_|\ \
    ____\_\  \ \_______\ \__\   \ \__\ \ \_______\
   |\_________\|_______|\|__|    \|__|  \|_______|
   \|_________|                    Version 1.8.26

Slite — independent rootless container engine (proot-based)
No distrobox, no podman, no daemon.
`

func main() {
	paths, err := config.Load()
	must(err)

	mgr := capsule.New(paths.Capsules, paths.Cache)

	// Best-effort sync with silverarch-repo's manifest on every run so new
	// distros show up automatically. Never fatal — falls back to defaults.
	_ = repo.Sync(paths.RepoCache)

	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "create":
		// slite create <distro-alias> <capsule-name>
		if len(args) < 3 {
			fatal("usage: slite create <distro> <capsule-name>")
		}
		must(mgr.Create(args[1], args[2]))
		fmt.Printf("✅ capsule %q created from %q\n", args[2], args[1])

	case "enter", "shell":
		if len(args) < 2 {
			fatal("usage: slite enter <capsule-name>")
		}
		must(mgr.Exec(args[1], nil))

	case "run", "exec":
		if len(args) < 3 {
			fatal("usage: slite run <capsule-name> <cmd...>")
		}
		must(mgr.Exec(args[1], args[2:]))

	case "install":
		if len(args) < 3 {
			fatal("usage: slite install <capsule-name> <package>")
		}
		pkgMgr, err := mgr.PkgMgr(args[1])
		must(err)
		cmdArgs, err := mgr.InstallCmd(pkgMgr, args[2])
		must(err)
		must(mgr.Exec(args[1], cmdArgs))

	case "list", "ls":
		names, err := mgr.List()
		must(err)
		if len(names) == 0 {
			fmt.Println("No capsules yet. Create one with: slite create <distro> <name>")
			return
		}
		for _, n := range names {
			fmt.Println(" •", n)
		}

	case "remove", "rm":
		if len(args) < 2 {
			fatal("usage: slite remove <capsule-name>")
		}
		must(mgr.Remove(args[1]))
		fmt.Printf("🗑  capsule %q removed\n", args[1])

	case "distros":
		printDistros()

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Printf("unknown command: %s\n\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(banner)
	fmt.Println("Usage:")
	fmt.Println("  slite create  <distro> <name>     Create a capsule from a base distro")
	fmt.Println("  slite enter   <name>              Enter a capsule shell")
	fmt.Println("  slite run     <name> <cmd...>      Run one command inside a capsule")
	fmt.Println("  slite install <name> <package>     Install a package inside a capsule")
	fmt.Println("  slite list                         List existing capsules")
	fmt.Println("  slite remove  <name>               Delete a capsule")
	fmt.Println("  slite distros                      List available base distros")
	fmt.Println()
	fmt.Println("Discord: https://discord.gg/eBgVHSPru9")
	fmt.Println("Wiki:    https://silverarchlinux.miraheze.org")
}

func printDistros() {
	aliases := make([]string, 0, len(rootfs.Registry))
	for a := range rootfs.Registry {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	fmt.Println("Available base distros:")
	for _, a := range aliases {
		e := rootfs.Registry[a]
		fmt.Printf("  %-8s %s (%s)\n", a, e.Name, e.PkgMgr)
	}
}

func must(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}
