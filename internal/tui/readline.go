package tui

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// ════════════════════════════════════════════════════════════════════
// Minimal raw-mode line editor with history — stdlib only, no deps.
// Supports: printable chars, backspace, up/down arrow history, Enter,
// Ctrl+C (returns ErrInterrupted), Ctrl+D on empty line (returns EOF).
// ════════════════════════════════════════════════════════════════════

type termios struct {
	Iflag, Oflag, Cflag, Lflag uint32
	Cc                         [20]byte
	Ispeed, Ospeed             uint32
}

const (
	tcgets = 0x5401
	tcsets = 0x5402
)

func getTermios(fd uintptr) (*termios, error) {
	var t termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, tcgets, uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return nil, errno
	}
	return &t, nil
}

func setTermios(fd uintptr, t *termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, tcsets, uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return errno
	}
	return nil
}

// ErrInterrupted is returned when the user presses Ctrl+C.
var ErrInterrupted = fmt.Errorf("interrupted")

// ErrEOF is returned when the user presses Ctrl+D on an empty line.
var ErrEOF = fmt.Errorf("eof")

// History holds past input lines for up/down arrow recall.
type History struct {
	lines []string
}

func NewHistory() *History { return &History{} }

func (h *History) Add(line string) {
	if line == "" {
		return
	}
	if len(h.lines) > 0 && h.lines[len(h.lines)-1] == line {
		return // don't duplicate consecutive identical entries
	}
	h.lines = append(h.lines, line)
}

// ReadLine reads one line with the given prompt, supporting Up/Down for
// history navigation. Falls back to a plain fmt.Scanln-style read if raw
// mode can't be enabled (e.g. stdin isn't a real TTY — piped input, CI).
func ReadLine(prompt string, hist *History) (string, error) {
	fd := os.Stdin.Fd()
	orig, err := getTermios(fd)
	if err != nil {
		// not a TTY (piped input) — fall back to simple line read
		return readLineFallback(prompt)
	}

	raw := *orig
	raw.Lflag &^= 0x8 | 0x2 // ECHO | ICANON off
	if err := setTermios(fd, &raw); err != nil {
		return readLineFallback(prompt)
	}
	defer setTermios(fd, orig)

	fmt.Print(prompt)

	var buf []rune
	histIdx := len(hist.lines) // one past the end = "not browsing yet"
	var savedCurrent string    // what the user was typing before pressing Up

	redraw := func() {
		fmt.Print("\r\033[K", prompt, string(buf))
	}

	readByte := func() (byte, error) {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		return b[0], err
	}

	for {
		b, err := readByte()
		if err != nil {
			return "", err
		}

		switch b {
		case '\r', '\n':
			fmt.Print("\r\n")
			return string(buf), nil

		case 3: // Ctrl+C
			fmt.Print("\r\n")
			return "", ErrInterrupted

		case 4: // Ctrl+D
			if len(buf) == 0 {
				fmt.Print("\r\n")
				return "", ErrEOF
			}

		case 127, 8: // Backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				redraw()
			}

		case 27: // ESC — start of an arrow-key sequence
			b1, _ := readByte()
			b2, _ := readByte()
			if b1 == '[' {
				switch b2 {
				case 'A': // Up
					if histIdx > 0 {
						if histIdx == len(hist.lines) {
							savedCurrent = string(buf)
						}
						histIdx--
						buf = []rune(hist.lines[histIdx])
						redraw()
					}
				case 'B': // Down
					if histIdx < len(hist.lines) {
						histIdx++
						if histIdx == len(hist.lines) {
							buf = []rune(savedCurrent)
						} else {
							buf = []rune(hist.lines[histIdx])
						}
						redraw()
					}
				}
			}

		default:
			if b >= 32 && b < 127 { // printable ASCII
				buf = append(buf, rune(b))
				fmt.Print(string(b))
			}
		}
	}
}

func readLineFallback(prompt string) (string, error) {
	fmt.Print(prompt)
	var line string
	_, err := fmt.Scanln(&line)
	if err != nil && err.Error() != "unexpected newline" {
		return "", err
	}
	return line, nil
}
