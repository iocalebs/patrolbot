package patrol

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iocalebs/patrolbot/internal/patroller"
	"github.com/spf13/cobra"
)

// CLIPatroller provides a command-line interface for patrolling revisions.
type CLIPatroller struct {
	cmd       *cobra.Command
	patroller *patroller.Patroller
	theme     Theme
}

// Theme defines the styles for the CLI output text.
type Theme struct {
	comment     lipgloss.Style
	error       lipgloss.Style
	hunk        lipgloss.Style
	lineAdded   lipgloss.Style
	lineRemoved lipgloss.Style
	prompt      lipgloss.Style
	title       lipgloss.Style
}

func newCLIPatroller(patroller *patroller.Patroller, cmd *cobra.Command, theme Theme) *CLIPatroller {
	return &CLIPatroller{
		cmd:       cmd,
		patroller: patroller,
		theme:     theme,
	}
}

func (p *CLIPatroller) run() error {
	cmd := p.cmd
	patroller := p.patroller
	scanner := bufio.NewScanner(cmd.InOrStdin())

	cmd.SetOut(os.Stdout)

	for patroller.Next(cmd.Context()) {
		if patroller.Err() != nil {
			return fmt.Errorf("error querying MediaWiki: %w", patroller.Err())
		}

		cmd.Println("\n>>> " + p.theme.title.Render(patroller.Title()) + " <<<")

		switch {
		case patroller.RevisionType() == "new":
			cmd.Println("New page")
		case patroller.LogAction() == "upload":
			cmd.Println("New file")
		case patroller.LogAction() == "overwrite":
			cmd.Println("New file version")
		default:
			cmd.Println("Edit")
		}

		cmd.Println("User:" + patroller.User())
		//nolint:gosmopolitan
		// This command runs on a user's local machine -- we want to display timestamps in their local timezone.
		cmd.Println(patroller.Timestamp().Local().Format(time.RFC1123))
		cmd.Println("Comment: " + p.theme.comment.Render(patroller.Comment()))

		if patroller.RevisionType() != "log" {
			diff, err := p.diff()
			if err != nil {
				p.printErr("Error generating diff: " + err.Error())
			} else {
				cmd.Println("\n" + diff + "\n")
			}
		}

		quit, err := p.prompt(scanner)
		if err != nil {
			return err
		}

		if quit {
			break
		}
	}

	return nil
}

func (p *CLIPatroller) prompt(scanner *bufio.Scanner) (bool, error) {
	cmd := p.cmd
	theme := p.theme

	for {
		cmd.Print(theme.prompt.Render("Mark as patrolled? ([y]es, [N]o, [t]hank, [o]pen in browser, [q]uit): "))

		if !scanner.Scan() {
			if scanner.Err() != nil {
				return true, fmt.Errorf("error reading input: %w", scanner.Err())
			}

			return true, nil
		}

		input := scanner.Text()
		switch strings.ToLower(input) {
		case "y":
			p.markPatrolled()
			return false, nil
		case "n", "":
			return false, nil
		case "t":
			p.thank()
		case "o":
			p.open()
		case "q":
			return true, nil
		default:
			p.printErr("\nUnknown command '" + input + "'")
		}
	}
}

func (p *CLIPatroller) diff() (string, error) {
	diff, err := p.patroller.Diff(p.cmd.Context())
	if err != nil {
		return "", fmt.Errorf("error generating revision diff: %w", err)
	}

	diff = strings.Replace(diff, `<tr><td colspan="4"><pre>`, "", 1)
	diff = strings.Replace(diff, "\n</pre></td></tr>", "", 1)
	diff = strings.ReplaceAll(diff, "\n \n ", "\n\n")
	diff = html.UnescapeString(diff)

	var sb strings.Builder

	for line := range strings.SplitSeq(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "@@"):
			sb.WriteString(p.theme.hunk.Render(line))
		case strings.HasPrefix(line, "+"):
			sb.WriteString(p.theme.lineAdded.Render(line))
		case strings.HasPrefix(line, "-"):
			sb.WriteString(p.theme.lineRemoved.Render(line))
		default:
			sb.WriteString(line)
		}

		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (p *CLIPatroller) markPatrolled() {
	err := p.patroller.MarkPatrolled(p.cmd.Context())
	if err != nil {
		p.printErr("Error marking revision as patrolled: " + err.Error())
	} else {
		p.cmd.Printf("Revision %d by %s marked as patrolled.\n", p.patroller.RevisionID(), p.patroller.User())
	}
}

func (p *CLIPatroller) thank() {
	err := p.patroller.Thank(p.cmd.Context())
	if err != nil {
		p.printErr("Error thanking user: " + err.Error())
	} else {
		p.cmd.Printf("Thanked %s.\n", p.patroller.User())
	}
}

func (p *CLIPatroller) open() {
	cmd := p.cmd
	patroller := p.patroller

	url, err := patroller.URL()
	if err != nil {
		p.printErr("Error generating diff URL: " + err.Error())
	}

	cmd.Println("Opening revision in browser...")

	// #nosec G204 -- url is trusted and not user-supplied
	switch runtime.GOOS {
	case "linux":
		err = exec.CommandContext(cmd.Context(), "xdg-open", url).Start()
	case "windows":
		err = exec.CommandContext(cmd.Context(), "rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.CommandContext(cmd.Context(), "open", url).Start()
	default:
		p.printErr("Error opening browser: unsupported platform: " + runtime.GOOS)
	}

	if err != nil {
		p.cmd.Println("Error opening browser: " + err.Error())
	}
}

func (p *CLIPatroller) printErr(msg string) {
	p.cmd.PrintErrln(p.theme.error.Render(msg))
}
