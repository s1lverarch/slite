package tui

import (
	"fmt"
	"strings"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner runs a label + animated frame until Stop() is called.
// Use for indeterminate work (extract, configure) — never for downloads
// where we have real byte counts (use ProgressBar for those instead).
type Spinner struct {
	label string
	color string
	stop  chan struct{}
	done  chan struct{}
}

func NewSpinner(label, color string) *Spinner {
	return &Spinner{label: label, color: color, stop: make(chan struct{}), done: make(chan struct{})}
}

func (s *Spinner) Start() {
	go func() {
		defer close(s.done)
		i := 0
		for {
			select {
			case <-s.stop:
				return
			default:
				fmt.Printf("\r%s%s%s %s", s.color, spinnerFrames[i%len(spinnerFrames)], "\033[0m", s.label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop halts the spinner and prints a final status line in its place.
func (s *Spinner) Stop(finalIcon, finalColor, finalMsg string) {
	close(s.stop)
	<-s.done
	fmt.Printf("\r\033[K%s%s\033[0m %s\n", finalColor, finalIcon, finalMsg)
}

// ProgressBar renders a real, byte-accurate download bar. Call Update on
// every callback tick; it redraws in place on the same terminal line.
type ProgressBar struct {
	label string
	color string
	width int
}

func NewProgressBar(label, color string) *ProgressBar {
	return &ProgressBar{label: label, color: color, width: 28}
}

// Update redraws the bar. If total is 0 (server didn't send Content-Length),
// falls back to showing raw bytes downloaded with a moving indicator instead
// of a fake percentage.
func (p *ProgressBar) Update(downloaded, total int64) {
	if total <= 0 {
		fmt.Printf("\r\033[K%s%s\033[0m %s  %s downloaded",
			p.color, "↓", p.label, humanBytes(downloaded))
		return
	}
	pct := float64(downloaded) / float64(total)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(p.width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)
	fmt.Printf("\r\033[K%s%s %s\033[0m %3.0f%%  %s / %s",
		p.color, bar, p.label, pct*100, humanBytes(downloaded), humanBytes(total))
}

// Done finalizes the bar with a checkmark and newline.
func (p *ProgressBar) Done(msg string) {
	fmt.Printf("\r\033[K\033[38;2;63;185;80m✓\033[0m %s\n", msg)
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
