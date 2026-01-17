package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/x/term"
	"github.com/iocalebs/patrolbot/internal/cli/clierr"
	"github.com/iocalebs/patrolbot/internal/report/reporter"
)

// A modification of [github.com/charmbracelet/fang.DefaultErrorHandler] that displays usage help
// when the command error is of type [UsageError].
// Fang's default error handler only looks for certain cobra-generated strings.
// See: https://github.com/spf13/cobra/pull/2266
func errHandler(w io.Writer, styles fang.Styles, err error) {
	if w, ok := w.(term.File); ok {
		// if stderr is not a tty, simply print the error without any
		// styling or going through an [ErrorHandler]:
		if !term.IsTerminal(w.Fd()) {
			_, _ = fmt.Fprintln(w, err.Error())
			return
		}
	}

	_, _ = fmt.Fprintln(w, styles.ErrorHeader.String())

	var multiError *clierr.MultiError
	if errors.As(err, &multiError) {
		style := styles.ErrorText.UnsetTransform().Transform(func(s string) string {
			if s == "" {
				return s
			}

			return strings.ToUpper(s[0:1]) + s[1:]
		})
		_, _ = fmt.Fprintln(w, style.Render(err.Error()))
	} else {
		_, _ = fmt.Fprintln(w, styles.ErrorText.Render(err.Error()+"."))
	}

	_, _ = fmt.Fprintln(w)
	if errors.Is(err, reporter.ErrReportTypeNotFound) {
		_, _ = fmt.Fprintln(w, lipgloss.JoinHorizontal(
			lipgloss.Left,
			styles.ErrorText.UnsetWidth().Render("Use"),
			styles.Program.Flag.Render(" --list "),
			styles.ErrorText.UnsetWidth().UnsetMargins().UnsetTransform().Render("to see available report types."),
		))
		_, _ = fmt.Fprintln(w)
	} else if isUsageError(err) {
		_, _ = fmt.Fprintln(w, lipgloss.JoinHorizontal(
			lipgloss.Left,
			styles.ErrorText.UnsetWidth().Render("Try"),
			styles.Program.Flag.Render(" --help "),
			styles.ErrorText.UnsetWidth().UnsetMargins().UnsetTransform().Render("for usage."),
		))
		_, _ = fmt.Fprintln(w)
	}
}

// Copied from https://github.com/charmbracelet/fang/blob/v0.4.4/fang.go
// with slight modification for custom error types.
func isUsageError(err error) bool {
	msg := err.Error()

	for _, prefix := range []string{
		"flag needs an argument:",
		"unknown flag:",
		"unknown shorthand flag:",
		"unknown command",
		"invalid argument",
	} {
		var errUsage clierr.UsageError
		if strings.HasPrefix(msg, prefix) || errors.As(err, &errUsage) {
			return true
		}
	}

	return false
}
