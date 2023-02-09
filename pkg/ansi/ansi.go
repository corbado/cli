package ansi

import (
	"fmt"
	"io"
	"os"

	"github.com/logrusorgru/aurora"
	"golang.org/x/term"
)

type Ansi struct {
	au aurora.Aurora
}

// New returns new ansi instance
func New(useColors bool, w io.Writer) *Ansi {
	if !isTerminal(w) {
		useColors = false
	}

	return &Ansi{
		au: aurora.NewAurora(useColors),
	}
}

// Bold makes given message bold
func (a *Ansi) Bold(message string) string {
	return a.au.Sprintf(a.au.Bold(message))
}

// Green colors given message green
func (a *Ansi) Green(message string) string {
	return a.au.Sprintf(a.au.Green(message))
}

// Red colors given message red
func (a *Ansi) Red(message string) string {
	return a.au.Sprintf(a.au.Red(message))
}

// ColorizeHTTPStatusCode colorizes given HTTP status code in green, yellow and red
func (a *Ansi) ColorizeHTTPStatusCode(httpStatusCode int) aurora.Value {
	status := fmt.Sprintf(" %d ", httpStatusCode)

	switch {
	case httpStatusCode >= 400:
		return a.au.BgRed(status).BrightWhite()

	case httpStatusCode >= 300:
		return a.au.BgYellow(status).BrightWhite()

	default:
		return a.au.BgGreen(status).BrightWhite()
	}
}

func isTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return term.IsTerminal(int(v.Fd()))

	default:
		return false
	}
}
