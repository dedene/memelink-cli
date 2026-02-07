// Package ui provides terminal output writers with color profile support.
package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// ErrInvalidColor is returned when an unsupported --color value is given.
var ErrInvalidColor = errors.New("invalid --color value")

const colorNever = "never"

// Options configures the UI.
type Options struct {
	Stdout io.Writer
	Stderr io.Writer
	Color  string // auto, always, never
}

// UI wraps stdout and stderr printers with color profile support.
type UI struct {
	out *Printer
	err *Printer
}

// New creates a UI with the given options.
func New(opts Options) (*UI, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}

	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	colorMode := strings.ToLower(strings.TrimSpace(opts.Color))
	if colorMode == "" {
		colorMode = "auto"
	}

	if colorMode != "auto" && colorMode != "always" && colorMode != colorNever {
		return nil, fmt.Errorf("%w: %q (expected auto|always|never)", ErrInvalidColor, colorMode)
	}

	out := termenv.NewOutput(opts.Stdout, termenv.WithProfile(termenv.EnvColorProfile()))
	errOut := termenv.NewOutput(opts.Stderr, termenv.WithProfile(termenv.EnvColorProfile()))

	outProfile := chooseProfile(out.Profile, colorMode)
	errProfile := chooseProfile(errOut.Profile, colorMode)

	return &UI{
		out: newPrinter(out, outProfile),
		err: newPrinter(errOut, errProfile),
	}, nil
}

// chooseProfile resolves the effective color profile from detected capability and user preference.
func chooseProfile(detected termenv.Profile, mode string) termenv.Profile {
	if termenv.EnvNoColor() {
		return termenv.Ascii
	}

	switch mode {
	case colorNever:
		return termenv.Ascii
	case "always":
		return termenv.TrueColor
	default:
		return detected
	}
}

// Out returns the stdout printer.
func (u *UI) Out() *Printer { return u.out }

// Err returns the stderr printer.
func (u *UI) Err() *Printer { return u.err }

// Printer wraps a termenv.Output with a resolved color profile.
type Printer struct {
	o       *termenv.Output
	profile termenv.Profile
}

func newPrinter(o *termenv.Output, profile termenv.Profile) *Printer {
	return &Printer{o: o, profile: profile}
}

// ColorEnabled returns true when color output is active.
func (p *Printer) ColorEnabled() bool { return p.profile != termenv.Ascii }

func (p *Printer) line(s string) {
	_, _ = io.WriteString(p.o, s+"\n")
}

func (p *Printer) printf(format string, args ...any) {
	p.line(fmt.Sprintf(format, args...))
}

// Print writes a string without a trailing newline.
func (p *Printer) Print(msg string) {
	_, _ = io.WriteString(p.o, msg)
}

// Println writes a line to the output.
func (p *Printer) Println(msg string) { p.line(msg) }

// Printf writes a formatted line to the output.
func (p *Printer) Printf(format string, args ...any) { p.printf(format, args...) }

// Errorf writes a formatted error line prefixed with "Error: ".
func (p *Printer) Errorf(format string, args ...any) {
	msg := fmt.Sprintf("Error: "+format, args...)
	if p.ColorEnabled() {
		msg = termenv.String(msg).Foreground(p.profile.Color("#ef4444")).String()
	}

	p.line(msg)
}

// Successf writes a formatted success line with green color.
func (p *Printer) Successf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if p.ColorEnabled() {
		msg = termenv.String(msg).Foreground(p.profile.Color("#22c55e")).String()
	}

	p.line(msg)
}

type uiCtxKey struct{}

// WithUI stores the UI in the context.
func WithUI(ctx context.Context, u *UI) context.Context {
	return context.WithValue(ctx, uiCtxKey{}, u)
}

// FromContext retrieves the UI from the context.
func FromContext(ctx context.Context) *UI {
	if v := ctx.Value(uiCtxKey{}); v != nil {
		if u, ok := v.(*UI); ok {
			return u
		}
	}

	return nil
}
