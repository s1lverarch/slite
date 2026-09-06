<div align="center">

<img src="https://readme-typing-svg.demolab.com/?font=JetBrains+Mono&weight=700&size=26&pause=1000&color=4DA6FF&center=true&vCenter=true&width=650&lines=Slite;Rootless+Container+Engine;No+distrobox.+No+podman.+No+daemon.;50+Linux+Distros%2C+One+CLI." alt="Typing SVG" />

<br>

[![Go Report Card](https://goreportcard.com/badge/github.com/s1lverarch/slite)](https://goreportcard.com/report/github.com/s1lverarch/slite)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue?style=for-the-badge&logo=gnu&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)](#)

<br>

![Distros](https://img.shields.io/badge/🐧_Base_Distros-6+-1793D1?style=for-the-badge&labelColor=0D3B6E)
![Engine](https://img.shields.io/badge/⚙️_Engine-proot-4DA6FF?style=for-the-badge&labelColor=0A2A50)
![Dependencies](https://img.shields.io/badge/📦_Daemon-None-3FB950?style=for-the-badge&labelColor=1A5C25)
![Root](https://img.shields.io/badge/🔓_Root_Required-No-E70013?style=for-the-badge&labelColor=8B0000)

</div>

---

## What is Slite?

**Slite** is an independent, rootless container engine built on [`proot`](https://proot-me.github.io/).
It lets you spin up isolated Linux distro environments — called **capsules** — from a single CLI,
without distrobox, without podman, and without a background daemon.

```
[slite]# create saur arch-box
[slite]# enter arch-box
[slite]# install arch-box neovim
```

That's the whole workflow.

---

## Features

<table>
<tr>
<td width="50%" valign="top">

### 🐧 Multi-Distro
Create capsules from Arch, Alpine, Debian, Ubuntu, Void, Fedora — with more
distros synced automatically from [`silverarch-repo`](https://github.com/s1lverarch/silverarch-repo)'s manifest.

### 🔓 Truly Rootless
Built on `proot` for syscall emulation. No `sudo`, no privileged daemon,
no kernel namespaces required. Works the same on a desktop or Termux.

### 💻 Interactive Shell
Run `slite` with no arguments and drop into a live `[slite]#` prompt —
history (↑/↓), real boot diagnostics, and colorized output out of the box.

</td>
<td width="50%" valign="top">

### 📊 Real Progress, Not Fake Spinners
Downloads show byte-accurate progress bars. Extraction and configuration
show live spinners. Nothing is padded or simulated — if a step is instant,
it *looks* instant.

### 🩺 Built-in Diagnostics
`slite doctor` checks `proot`, `curl`, `tar`, directory permissions, and
capsule state — six real checks, not a placeholder.

### 🎨 Nerd Font Native
Every status line, error, and menu item uses consistent Nerd Font
iconography — install a [Nerd Font](https://www.nerdfonts.com/) once
and every corner of Slite lights up.

</td>
</tr>
</table>

---

## Install

### Arch Linux (via `silverarch-repo`)

```bash
git clone https://github.com/s1lverarch/slite.git
cd slite
makepkg -si
```

### From source

```bash
git clone https://github.com/s1lverarch/slite.git
cd slite
go build -o slite ./cmd/slite
sudo install -Dm755 slite /usr/bin/slite
```

**Runtime dependencies:** `proot`, `curl`, `tar`

```bash
sudo pacman -S proot curl tar
```

> `proot` isn't in Arch's official repos — grab it from the AUR:
> ```bash
> git clone https://aur.archlinux.org/proot.git && cd proot && makepkg -si
> ```

---

## Usage

### Interactive shell (default)

```bash
$ slite
```

<div align="center">

```
███████╗██╗     ██╗████████╗███████╗
██╔════╝██║     ██║╚══██╔══╝██╔════╝
███████╗██║     ██║   ██║   █████╗
╚════██║██║     ██║   ██║   ██╔══╝
███████║███████╗██║   ██║   ███████╗
╚══════╝╚══════╝╚═╝   ╚═╝   ╚══════╝
                      Version 1.9.4

 Independent rootless container engine (proot-based)
 No distrobox, no podman, no daemon.

[15:37:43]  Resolving home directory          OK (/home/user/.slite)
[15:37:43]  Checking capsules directory       OK (/home/user/.slite/capsules)
[15:37:43]  Checking cache directory          OK (/home/user/.slite/cache)
[15:37:43]  Checking proot availability       OK (/usr/bin/proot)
[15:37:43]  Checking curl availability        OK (/usr/bin/curl)
[15:37:43]  Syncing silverarch-repo manifest  OK (manifest up to date)
[15:37:44]  Loading capsules                  OK (2 found)

 Status:
   Version:            1.9.4
   Home:               /home/user/.slite
   Capsules:           2
   Distros available:  6

 Links:
   Wiki:     https://silverarchlinux.miraheze.org
   Discord:  https://discord.gg/eBgVHSPru9
   Source:   https://github.com/s1lverarch/slite

Type 'help' for commands, 'exit' to quit. (↑/↓ for history)
[slite]#
```

</div>

### One-shot / scripting mode

Every command also works directly from the shell for automation:

```bash
slite create saur arch-box       # create a capsule
slite enter arch-box              # drop into it
slite install arch-box neovim     # install a package
slite list                        # list capsules
slite distros                     # list available base distros
slite doctor                      # run diagnostics
slite update                      # check for a newer release
```

---

## Commands

| Command | Description |
|---|---|
| `create  <distro> <name>` | Create a capsule from a base distro |
| `enter   <name>` | Enter a capsule shell |
| `run     <name> <cmd...>` | Run one command inside a capsule |
| `install <name> <package>` | Install a package inside a capsule |
| `list` | List existing capsules |
| `remove  <name>` | Delete a capsule |
| `distros` | List available base distros |
| `status` | Show current install status |
| `info` | Show system + environment info |
| `version` | Show Slite version |
| `update` | Check for a newer release |
| `doctor` | Diagnose common setup problems |
| `clear` | Clear the screen |
| `exit` | Leave the shell |

---

## Supported Base Distros

| Alias | Distro | Package Manager |
|:-----:|--------|:----------------:|
| `saur` | Arch Linux | `pacman` |
| `salp` | Alpine Linux | `apk` |
| `sdur` | Debian | `apt` |
| `subu` | Ubuntu | `apt` |
| `svoid` | Void Linux | `xbps` |
| `sfur` | Fedora | `dnf` |

More distros sync automatically from [`silverarch-repo`](https://github.com/s1lverarch/silverarch-repo)'s manifest — no client update needed.

---

## Architecture

```
slite/
├── cmd/slite/main.go          # CLI entry point + interactive shell
└── internal/
    ├── capsule/                # capsule lifecycle (create, enter, install, remove)
    ├── config/                 # ~/.slite path resolution
    ├── repo/                   # silverarch-repo manifest sync + release checks
    ├── rootfs/                 # base distro registry + download logic
    └── tui/
        ├── readline.go          # raw-mode line editor with history (stdlib only)
        └── progress.go          # spinners + byte-accurate progress bars
```

No external Go dependencies — the entire TUI layer (history, spinners,
progress bars) is built on the standard library only.

---

## Why not distrobox / podman?

| | Slite | distrobox | podman |
|---|:---:|:---:|:---:|
| Daemon required | ❌ | ❌ | ✅ (rootless mode still needs `conmon`) |
| Root required | ❌ | ❌ | ⚠️ (depends on setup) |
| Works in Termux/proot chains | ✅ | ⚠️ | ❌ |
| Single static binary | ✅ | ❌ | ❌ |
| Dependencies | `proot`, `curl`, `tar` | `podman`/`docker` | full container runtime |

Slite trades container-runtime features (networking namespaces, cgroups,
image layering) for simplicity: it's a rootfs + `proot`, nothing more.

---

## Contributing

Issues and PRs welcome. This project follows the SilverArch Linux
open-source philosophy — build it minimal, ship it open.

```bash
git clone https://github.com/s1lverarch/slite.git
cd slite
go build ./...
go vet ./...
```

---

<div align="center">

### Part of the SilverArch Linux Project

[![Wiki](https://img.shields.io/badge/Wiki-4DA6FF?style=for-the-badge&logo=wikidot&logoColor=white)](https://silverarchlinux.miraheze.org)
[![Discord](https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/eBgVHSPru9)
[![GitHub](https://img.shields.io/badge/GitHub-181717?style=for-the-badge&logo=github&logoColor=white)](https://github.com/s1lverarch)

<br>

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=1,5,12,20&height=120&section=footer&text=&fontColor=fff" width="100%" />

**Made with 💙 in Tunis 🇹🇳**

</div>
