// Package init implements the `patrolbot config init` command
package init

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

//go:embed config.yaml.tmpl
var templateFS embed.FS

// NewCommand returns the `config init` Cobra command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize PatrolBot configuration file",
		Long:  "Interactively create an initial configuration file at $HOME/.patrolbot/config.yaml.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to obtain user home directory: %w", err)
			}

			path := filepath.Join(home, ".patrolbot", "config.yaml")

			_, err = os.Stat(path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("error checking home config file: %w", err)
			}

			exists := err == nil
			if exists {
				overwrite, err := promptOverwrite(cmd.Context(), path)
				if err != nil {
					return err
				}

				if !overwrite {
					return nil
				}
			}

			cfg, err := promptConfig(cmd.Context())
			if err != nil {
				return err
			}

			err = writeConfig(cfg, path)
			if err != nil {
				return err
			}

			cmd.Println("Wrote configuration file: " + path)

			return nil
		},
	}
}

func writeConfig(cfg config.Config, path string) error {
	tmpl, err := template.ParseFS(templateFS, "config.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing config template: %w", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return fmt.Errorf("error creating config.yaml from template: %w", err)
	}

	dir := filepath.Dir(path)

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	err = os.WriteFile(path, buf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("error writing config to file: %w", err)
	}

	return nil
}
