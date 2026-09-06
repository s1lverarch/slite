package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/s1lverarch/slite/internal/capsule"
	"github.com/s1lverarch/slite/internal/config"
	"github.com/s1lverarch/slite/internal/repo"
	"github.com/s1lverarch/slite/internal/rootfs"
	"github.com/s1lverarch/slite/internal/tui"
)

// ── ANSI colors ──────────────────────────────────────────────────────
const (
	nc     = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[38;2;77;166;255m"
	green  = "\033[38;2;63;185;80m"
	red    = "\033[38;2;255;90;90m"
	yellow = "\033[38;2;230;200;80m"
	gray   = "\033[38;2;90;110;140m"
)

// ── Nerd Font icons (require a Nerd Font in the terminal) ────────────
const (
	iconOS          = "\uf17c" // linux
	iconArch        = "\uf303" // arch
	iconHome        = "\uf015" // home
	iconFolder      = "\uf07c" // open folder
	iconCache       = "\uf1c0" // db/cache
	iconBox         = "\uf466" // box / container
	iconNetwork     = "\uf6ff" // network/globe
	iconCheck       = "\uf00c" // check
	iconCross       = "\uf00d" // cross
	iconWarn        = "\uf071" // warning
	iconGear        = "\uf013" // gear
	iconPackage     = "\uf487" // package
	iconDownload    = "\uf019" // download
	iconTerminal    = "\uf120" // terminal
	iconList        = "\uf03a" // list
	iconTrash       = "\uf1f8" // trash
	iconInfo        = "\uf05a" // info
	iconBook        = "\uf02d" // book/wiki
	iconChat        = "\uf1d7" // chat bubble
	iconGithub      = "\uf09b" // github
	iconBolt        = "\uf0e7" // bolt/version
	iconStethoscope = "\uf0f1" // doctor/health-check
	iconRefresh     = "\uf021" // refresh/update
)

var version = "1.9.4"

var banner = cyan + bold + `
███████╗██╗     ██╗████████╗███████╗
██╔════╝██║     ██║╚══██╔══╝██╔════╝
███████╗██║     ██║   ██║   █████╗
╚════██║██║     ██║   ██║   ██╔══╝
███████║███████╗██║   ██║   ███████╗
╚══════╝╚══════╝╚═╝   ╚═╝   ╚══════╝` + nc + dim + `
                      Version ` + version + nc + `

` + gray + iconOS + `  Independent rootless container engine (proot-based)
` + iconBox + `  No distrobox, no podman, no daemon.` + nc + `
`

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	paths, err := config.Load()
	must(err)

	mgr := capsule.New(paths.Capsules, paths.Cache)

	args := os.Args[1:]

	// one-shot mode: `slite <command> ...` still works exactly like before,
	// for scripting/automation — only bare `slite` drops into the shell.
	if len(args) > 0 {
		runCommand(mgr, paths, args)
		return
	}

	runInteractiveShell(mgr, paths)
}

// ════════════════════════════════════════════════════════════════════
// INTERACTIVE SHELL
// ════════════════════════════════════════════════════════════════════

func runInteractiveShell(mgr *capsule.Manager, paths *config.Paths) {
	clearScreen()
	fmt.Print(banner)
	fmt.Println()
	printBootLog(mgr, paths)
	fmt.Println()
	printStatusBlock(mgr, paths)
	fmt.Println()

	hist := tui.NewHistory()
	for {
		line, err := tui.ReadLine(cyan+"[slite]"+nc+"# ", hist)
		if err == tui.ErrEOF {
			fmt.Println()
			return
		}
		if err == tui.ErrInterrupted {
			continue // Ctrl+C just cancels the current line, like a real shell
		}
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		hist.Add(line)
		args := strings.Fields(line)

		switch args[0] {
		case "exit", "quit":
			fmt.Println(dim + "Goodbye." + nc)
			return
		case "clear":
			clearScreen()
			fmt.Print(banner)
		default:
			runCommand(mgr, paths, args)
		}
	}
}

// printBootLog shows REAL state — no fake kernel spam. Every line here
// reflects something actually true about this machine/install right now.
func printBootLog(mgr *capsule.Manager, paths *config.Paths) {
	steps := []struct {
		icon  string
		label string
		fn    func() (string, bool)
	}{
		{iconHome, "Resolving home directory", func() (string, bool) {
			return paths.Home, paths.Home != ""
		}},
		{iconFolder, "Checking capsules directory", func() (string, bool) {
			return paths.Capsules, dirExists(paths.Capsules)
		}},
		{iconCache, "Checking cache directory", func() (string, bool) {
			return paths.Cache, dirExists(paths.Cache)
		}},
		{iconBox, "Checking proot availability", func() (string, bool) {
			return whichOrMissing("proot")
		}},
		{iconNetwork, "Checking curl availability", func() (string, bool) {
			return whichOrMissing("curl")
		}},
		{iconRefresh, "Syncing silverarch-repo manifest", func() (string, bool) {
			err := repo.Sync(paths.RepoCache)
			if err != nil {
				return "offline — using built-in defaults", true // not fatal, still "ok"
			}
			return "manifest up to date", true
		}},
		{iconList, "Loading capsules", func() (string, bool) {
			names, err := mgr.List()
			if err != nil {
				return "none found", true
			}
			return fmt.Sprintf("%d found", len(names)), true
		}},
	}

	for _, s := range steps {
		ts := time.Now().Format("15:04:05")
		detail, ok := s.fn()
		status := green + iconCheck + " OK" + nc
		if !ok {
			status = red + iconCross + " MISSING" + nc
		}
		fmt.Printf(gray+"[%s]"+nc+" %s %-30s %s"+dim+" (%s)"+nc+"\n", ts, s.icon, s.label, status, detail)
		time.Sleep(35 * time.Millisecond) // tiny, real, not padded — feels alive without being fake
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func whichOrMissing(bin string) (string, bool) {
	path, err := exec.LookPath(bin)
	if err != nil {
		return "not found in PATH", false
	}
	return path, true
}

// printStatusBlock — the real, current, queryable state of this install.
func printStatusBlock(mgr *capsule.Manager, paths *config.Paths) {
	names, _ := mgr.List()
	fmt.Println(bold + iconGear + " Status:" + nc)
	fmt.Printf("  %s %-16s %s\n", iconBolt, "Version:", version)
	fmt.Printf("  %s %-16s %s\n", iconHome, "Home:", paths.Home)
	fmt.Printf("  %s %-16s %d\n", iconBox, "Capsules:", len(names))
	fmt.Printf("  %s %-16s %s\n", iconArch, "Distros available:", fmt.Sprint(len(rootfs.Registry)))
	fmt.Println()
	fmt.Println(bold + iconBook + " Links:" + nc)
	fmt.Printf("  %s %-16s "+cyan+"%s"+nc+"\n", iconBook, "Wiki:", "https://silverarchlinux.miraheze.org")
	fmt.Printf("  %s %-16s "+cyan+"%s"+nc+"\n", iconChat, "Discord:", "https://discord.gg/eBgVHSPru9")
	fmt.Printf("  %s %-16s "+cyan+"%s"+nc+"\n", iconGithub, "Source:", "https://github.com/s1lverarch/slite")
	fmt.Println()
	fmt.Println(dim + "Type 'help' for commands, 'exit' to quit. (↑/↓ for history)" + nc)
}

// ════════════════════════════════════════════════════════════════════
// COMMAND DISPATCH — shared by shell mode and one-shot CLI mode
// ════════════════════════════════════════════════════════════════════

func runCommand(mgr *capsule.Manager, paths *config.Paths, args []string) {
	switch args[0] {
	case "create":
		if len(args) < 3 {
			warn("usage: create <distro> <capsule-name>")
			return
		}
		runCreate(mgr, args[1], args[2])

	case "enter", "shell":
		if len(args) < 2 {
			warn("usage: enter <capsule-name>")
			return
		}
		must(mgr.Exec(args[1], nil))

	case "run", "exec":
		if len(args) < 3 {
			warn("usage: run <capsule-name> <cmd...>")
			return
		}
		must(mgr.Exec(args[1], args[2:]))

	case "install":
		if len(args) < 3 {
			warn("usage: install <capsule-name> <package>")
			return
		}
		runInstall(mgr, args[1], args[2])

	case "list", "ls":
		names, err := mgr.List()
		if err != nil || len(names) == 0 {
			fmt.Println(dim + "No capsules yet. Create one with: create <distro> <capsule-name>" + nc)
			return
		}
		for _, n := range names {
			fmt.Println(" " + cyan + iconBox + nc + " " + n)
		}

	case "remove", "rm":
		if len(args) < 2 {
			warn("usage: remove <capsule-name>")
			return
		}
		if err := mgr.Remove(args[1]); err != nil {
			warn(err.Error())
			return
		}
		okMsg(iconTrash + " capsule " + fmt.Sprintf("%q", args[1]) + " removed")

	case "distros":
		printDistros()

	case "status":
		printStatusBlock(mgr, paths)

	case "info":
		printInfo()

	case "version", "-v", "--version":
		fmt.Println(cyan + iconBolt + nc + " slite " + version)

	case "update":
		runUpdate(paths)

	case "doctor":
		runDoctor(mgr, paths)

	case "help", "-h", "--help":
		printHelp()

	default:
		warn(fmt.Sprintf("unknown command: %s (try 'help')", args[0]))
	}
}

// ── create: real progress bar for download, spinner for extract/configure
func runCreate(mgr *capsule.Manager, alias, name string) {
	var bar *tui.ProgressBar
	var spin *tui.Spinner

	err := mgr.CreateWithProgress(alias, name, func(phase string, downloaded, total int64) {
		switch phase {
		case "download":
			if bar == nil {
				bar = tui.NewProgressBar(fmt.Sprintf("%s %s", iconDownload, alias), cyan)
			}
			bar.Update(downloaded, total)
		case "extract":
			if bar != nil {
				bar.Done(fmt.Sprintf("%s downloaded", alias))
				bar = nil
			}
			spin = tui.NewSpinner(iconGear+" extracting rootfs...", cyan)
			spin.Start()
		case "configure":
			if spin != nil {
				spin.Stop(iconCheck, green, "extraction complete")
			}
			spin = tui.NewSpinner(iconGear+" configuring capsule...", cyan)
			spin.Start()
		}
	})

	if spin != nil {
		if err != nil {
			spin.Stop(iconCross, red, "failed: "+err.Error())
		} else {
			spin.Stop(iconCheck, green, "capsule configured")
		}
	}

	if err != nil {
		warn(err.Error())
		return
	}
	okMsg(fmt.Sprintf("%s capsule %q created from %q", iconBox, name, alias))
}

// ── install: spinner while the package manager runs (no byte-progress
// available here since it's an interactive subprocess, not a plain download)
func runInstall(mgr *capsule.Manager, name, pkg string) {
	pkgMgr, err := mgr.PkgMgr(name)
	if err != nil {
		warn(err.Error())
		return
	}
	cmdArgs, err := mgr.InstallCmd(pkgMgr, pkg)
	if err != nil {
		warn(err.Error())
		return
	}
	fmt.Printf(gray+"%s running %s install for %s...\n"+nc, iconPackage, pkgMgr, pkg)
	must(mgr.Exec(name, cmdArgs))
}

func okMsg(msg string) { fmt.Println(green + iconCheck + " " + nc + msg) }
func warn(msg string)  { fmt.Println(red + iconCross + " " + nc + msg) }

func printHelp() {
	fmt.Println(bold + iconInfo + " Commands:" + nc)
	rows := [][3]string{
		{iconBox, "create  <distro> <name>", "Create a capsule from a base distro"},
		{iconTerminal, "enter   <name>", "Enter a capsule shell"},
		{iconTerminal, "run     <name> <cmd...>", "Run one command inside a capsule"},
		{iconPackage, "install <name> <package>", "Install a package inside a capsule"},
		{iconList, "list", "List existing capsules"},
		{iconTrash, "remove  <name>", "Delete a capsule"},
		{iconArch, "distros", "List available base distros"},
		{iconGear, "status", "Show current install status"},
		{iconInfo, "info", "Show system + environment info"},
		{iconBolt, "version", "Show slite version"},
		{iconRefresh, "update", "Check for a newer slite release"},
		{iconStethoscope, "doctor", "Diagnose common setup problems"},
		{iconTerminal, "clear", "Clear the screen"},
		{iconCross, "exit", "Leave the shell"},
	}
	for _, r := range rows {
		fmt.Printf("  %s "+cyan+"%-28s"+nc+" %s\n", r[0], r[1], r[2])
	}
	fmt.Println()
	fmt.Println(bold + iconBook + " Docs:" + nc + "    " + cyan + "https://silverarchlinux.miraheze.org" + nc)
	fmt.Println(bold + iconChat + " Discord:" + nc + " " + cyan + "https://discord.gg/eBgVHSPru9" + nc)
	fmt.Println(bold + iconGithub + " Source:" + nc + "  " + cyan + "https://github.com/s1lverarch/slite" + nc)
}

func printDistros() {
	aliases := make([]string, 0, len(rootfs.Registry))
	for a := range rootfs.Registry {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	fmt.Println(bold + iconArch + " Available base distros:" + nc)
	for _, a := range aliases {
		e := rootfs.Registry[a]
		fmt.Printf("  %s "+cyan+"%-8s"+nc+" %s "+dim+"(%s)"+nc+"\n", iconBox, a, e.Name, e.PkgMgr)
	}
}

// ── info: environment snapshot, useful for bug reports
func printInfo() {
	fmt.Println(bold + iconInfo + " Environment:" + nc)
	fmt.Printf("  %s %-16s %s\n", iconBolt, "slite version:", version)
	fmt.Printf("  %s %-16s %s\n", iconGear, "Go runtime:", runtime.Version())
	fmt.Printf("  %s %-16s %s/%s\n", iconOS, "OS/Arch:", runtime.GOOS, runtime.GOARCH)
	if home, err := os.UserHomeDir(); err == nil {
		fmt.Printf("  %s %-16s %s\n", iconHome, "Home:", home)
	}
	prootPath, prootOk := whichOrMissing("proot")
	status := green + iconCheck + nc
	if !prootOk {
		status = red + iconCross + nc
	}
	fmt.Printf("  %s %-16s %s %s\n", iconBox, "proot:", status, prootPath)
}

// ── update: checks GitHub releases (best-effort, never fatal)
func runUpdate(paths *config.Paths) {
	spin := tui.NewSpinner(iconRefresh+" checking for updates...", cyan)
	spin.Start()

	latest, err := repo.LatestVersion()
	if err != nil {
		spin.Stop(iconWarn, yellow, "couldn't check for updates (offline?)")
		return
	}

	if latest == version {
		spin.Stop(iconCheck, green, "you're on the latest version ("+version+")")
		return
	}
	spin.Stop(iconBolt, cyan, "update available: "+version+" → "+latest)
	fmt.Println(dim + "Run your package manager to upgrade, e.g.: sudo pacman -Syu slite" + nc)
}

// ── doctor: diagnoses common problems, like a health check
func runDoctor(mgr *capsule.Manager, paths *config.Paths) {
	fmt.Println(bold + iconStethoscope + " Running diagnostics:" + nc)
	fmt.Println()

	checks := []struct {
		label string
		fn    func() (bool, string)
	}{
		{"proot installed", func() (bool, string) {
			p, ok := whichOrMissing("proot")
			return ok, p
		}},
		{"curl installed", func() (bool, string) {
			p, ok := whichOrMissing("curl")
			return ok, p
		}},
		{"tar installed", func() (bool, string) {
			p, ok := whichOrMissing("tar")
			return ok, p
		}},
		{"home directory writable", func() (bool, string) {
			testFile := paths.Home + "/.slite-write-test"
			err := os.WriteFile(testFile, []byte("ok"), 0o644)
			if err != nil {
				return false, err.Error()
			}
			os.Remove(testFile)
			return true, paths.Home
		}},
		{"capsules directory exists", func() (bool, string) {
			return dirExists(paths.Capsules), paths.Capsules
		}},
		{"at least one capsule created", func() (bool, string) {
			names, _ := mgr.List()
			if len(names) == 0 {
				return false, "none yet — try: create saur mybox"
			}
			return true, fmt.Sprintf("%d capsule(s)", len(names))
		}},
	}

	failures := 0
	for _, c := range checks {
		ok, detail := c.fn()
		if ok {
			fmt.Printf("  %s%s%s %s "+dim+"(%s)"+nc+"\n", green, iconCheck, nc, c.label, detail)
		} else {
			failures++
			fmt.Printf("  %s%s%s %s "+dim+"(%s)"+nc+"\n", red, iconCross, nc, c.label, detail)
		}
	}

	fmt.Println()
	if failures == 0 {
		fmt.Println(green + iconCheck + " All checks passed." + nc)
	} else {
		fmt.Printf(yellow+"%s %d check(s) need attention.\n"+nc, iconWarn, failures)
	}
}

func must(err error) {
	if err != nil {
		warn(err.Error())
	}
}
