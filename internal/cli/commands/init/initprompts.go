package init

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iocalebs/patrolbot/internal/config"
)

// Capitalized user-facing error message.
var errInvalidInput = errors.New("Invalid input") //nolint:staticcheck

func notEmpty(fieldName string) func(string) error {
	return func(fieldValue string) error {
		if fieldValue == "" {
			return fmt.Errorf("%w: %s cannot be empty", errInvalidInput, fieldName)
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
	)

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
		url  url.URL
	)

	wiki.Site.ScriptPath = "/w"
	wiki.Site.ArticlePath = "/wiki"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Wiki URL").
				Placeholder("https://en.wikipedia.org").
				Value(&wiki.Site.URL).
				Validate(func(input string) error {
					if input == "" {
						return fmt.Errorf("%w: URL cannot be empty", errInvalidInput)
					}

					u, err := url.Parse(input)
					if err != nil {
						// Capitalized user-facing error message.
						return fmt.Errorf("Invalid URL: %w", err) //nolint:staticcheck
					}

					url = *u

					return nil
				}),

			huh.NewInput().
				Title("Wiki name").
				Description("A short, one-word ID for use in commands.").
				Value(&name).
				Validate(notEmpty("Wiki name")),

			huh.NewInput().
				Title("Script path").
				Description("You can confirm the wiki's script path by adding {{SCRIPTPATH}} to a sandbox page.").
				Value(&wiki.Site.ScriptPath),

			huh.NewInput().
				Title("Article path").
				Description(
					"You can confirm the wiki's article path by adding {{ARTICLEPATH}} to a sandbox page. "+
						"Omit the /$1 at the end.",
				).
				Value(&wiki.Site.ArticlePath),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Create a bot password for PatrolBot").
				DescriptionFunc(func() string {
					botPasswordsURL := url.JoinPath(wiki.Site.ArticlePath, "Special:BotPasswords")

					return `  1. Go to ` + botPasswordsURL.String() + `
  2. Create a new bot password with bot name "PatrolBot"
  3. Grant "Patrol changes to pages", then click "Create"
  4. Enter bot username and password below.`
				}, []*string{&wiki.Site.URL, &wiki.Site.ArticlePath}),

			huh.NewInput().
				Title("Username").
				Placeholder("WikiUsername@PatrolBot").
				Value(&wiki.Auth.Username),

			huh.NewInput().
				Title("Password").
				Value(&wiki.Auth.Password),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Contact email address").
				Description(
					"Please provide an email address to be sent in MediaWiki API requests. "+
						"This allows the wiki's system administrator to contact you if there's an issue with the "+
						"bot's API usage",
				).
				Value(&wiki.Client.From),
		),
	)

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
