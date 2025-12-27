package init

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/mattn/go-isatty"
)

func notEmpty(fieldName string) func(string) error {
	return func(fieldValue string) error {
		if fieldValue == "" {
			return fmt.Errorf("%s cannot be empty", fieldName) //nolint:err113
		}

		return nil
	}
}

func promptOverwrite(ctx context.Context, path string) (bool, error) {
	var overwrite bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Configuration file already exists: " + path + "\n\nOverwrite existing config?").
				WithButtonAlignment(lipgloss.Left).
				Value(&overwrite),
		),
	).WithAccessible(!isatty.IsTerminal(os.Stdout.Fd()))

	err := form.RunWithContext(ctx)
	if err != nil {
		return false, fmt.Errorf("prompt failed: %w", err)
	}

	if !overwrite {
		return false, nil
	}

	return true, nil
}

func promptConfig(ctx context.Context) (config.Config, error) { //nolint:funlen
	var (
		wiki config.Wiki
		name string
	)

	wiki.ScriptPath = "/w"
	wiki.ArticlePath = "/wiki"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Wiki URL").
				Placeholder("https://en.wikipedia.org").
				Value(&wiki.URL).
				Validate(notEmpty("URL")),

			huh.NewInput().
				Title("Wiki name").
				Description("A short, one-word ID for use in commands.").
				Value(&name).
				Validate(notEmpty("Wiki name")),

			huh.NewInput().
				Title("Script path").
				Description("You can confirm the wiki's script path by adding {{SCRIPTPATH}} to a sandbox page.").
				Value(&wiki.ScriptPath).
				Validate(notEmpty("Script path")),

			huh.NewInput().
				Title("Article path").
				Description(
					"You can confirm the wiki's article path by adding {{ARTICLEPATH}} to a sandbox page. "+
						"Omit the /$1 at the end.",
				).
				Value(&wiki.ArticlePath).
				Validate(notEmpty("Article path")),
		),

		huh.NewGroup(
			huh.NewInput().
				TitleFunc(func() string {
					return "Create a bot password for patrolbot at " + wiki.URL + wiki.ArticlePath +
						"/Special:BotPasswords\n\nUsername"
				}, []*string{&wiki.URL, &wiki.ArticlePath}).
				Placeholder("WikiUser@BotPasswordName").
				Value(&wiki.Username),

			huh.NewInput().
				Title("Password").
				Value(&wiki.Password),
		),
	).WithAccessible(!isatty.IsTerminal(os.Stdout.Fd()))

	err := form.RunWithContext(ctx)
	if err != nil {
		return config.Config{}, fmt.Errorf("prompt failed: %w", err)
	}

	cfg := config.Config{
		Wiki: name,
		Wikis: map[string]config.Wiki{
			name: wiki,
		},
	}

	return cfg, nil
}
