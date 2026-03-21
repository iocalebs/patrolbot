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

type formData struct {
	WikiName         string
	Wiki             config.Wiki
	Discord          config.Discord
	DiscordChannelID string
	DiscordServerID  string
}

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
				Title("Configuration already exists at " + path + "\n\nOverwrite existing config?").
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

func promptConfig(ctx context.Context) (formData, error) {
	var (
		data formData
		url  url.URL
	)

	data.Wiki.Site.ScriptPath = "/w"
	data.Wiki.Site.ArticlePath = "/wiki"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Wiki URL").
				Placeholder("https://en.wikipedia.org").
				Value(&data.Wiki.Site.URL).
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
				Value(&data.WikiName).
				Validate(notEmpty("Wiki name")),

			huh.NewInput().
				Title("Script path").
				Description("You can confirm the wiki's script path by adding {{SCRIPTPATH}} to a sandbox page.").
				Value(&data.Wiki.Site.ScriptPath),

			huh.NewInput().
				Title("Article path").
				Description(
					"You can confirm the wiki's article path by adding {{ARTICLEPATH}} to a sandbox page. "+
						"Omit the /$1 at the end.",
				).
				Value(&data.Wiki.Site.ArticlePath),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Create a bot password for PatrolBot").
				DescriptionFunc(func() string {
					botPasswordsURL := url.JoinPath(data.Wiki.Site.ArticlePath, "Special:BotPasswords")

					return `  1. Go to ` + botPasswordsURL.String() + `
  2. Create a new bot password with bot name "PatrolBot"
  3. Grant "Patrol changes to pages", then click "Create"
  4. Enter bot username and password below.`
				}, []*string{&data.Wiki.Site.URL, &data.Wiki.Site.ArticlePath}),

			huh.NewInput().
				Title("Username").
				Placeholder("WikiUsername@PatrolBot").
				Value(&data.Wiki.Auth.Username),

			huh.NewInput().
				Title("Password").
				Value(&data.Wiki.Auth.Password),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Contact email address").
				Description(
					"Please provide an email address to be sent in MediaWiki API requests. "+
						"This allows the wiki's system administrator to contact you if there's an issue with the "+
						"bot's API usage",
				).
				Value(&data.Wiki.Client.From),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Create a Discord app for PatrolBot").
				Description("https://discord.com/developers/applications?new\\_application=true"),

			huh.NewInput().
				Title("Application ID").
				Description("On the General Information page, copy the value for Application ID.").
				Value(&data.Discord.AppID),

			huh.NewInput().
				Title("Public key").
				Description("On the General Information page, copy the value for Public Key.").
				Value(&data.Discord.PublicKey),

			huh.NewInput().
				Title("Bot token").
				Description("On the Bot page under Token, click \"Reset Token\" to generate a new bot token.").
				Value(&data.Discord.Token),

			huh.NewNote().
				Title("Add bot to Discord server").
				Description(
					"On the Installation page under Install Link, open the Discord Provided Link in a browser "+
						"and add the bot to the desired server."),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Discord server ID").
				Description(
					"Enter the ID of the Discord server you added the bot to.\n"+
						"To obtain a server's ID, right click the server icon and click \"Copy Server ID\".").
				Value(&data.DiscordServerID),

			huh.NewInput().
				Title("Discord channel ID").
				Description(
					"Enter the ID of the Discord channel where PatrolBot should post reports.\n"+
						"To obtain a channel's ID, right click it in the channel list and click \"Copy Channel ID\".").
				Value(&data.DiscordChannelID),
		),
	)

	err := form.RunWithContext(ctx)
	if err != nil {
		return formData{}, fmt.Errorf("prompt failed: %w", err)
	}

	return data, nil
}
