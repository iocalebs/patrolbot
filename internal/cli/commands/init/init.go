// Package init implements the "patrolbot init" command
package init

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	//go:embed config.yaml.tmpl
	configTemplateFS embed.FS

	//go:embed templates/*
	templateDirFS embed.FS
)

// NewCommand returns the `config init` Cobra command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize PatrolBot configuration",
		Long:  "Interactively create an initial configuration file at $HOME/.patrolbot/config.yaml.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to obtain user home directory: %w", err)
			}

			dir := filepath.Join(home, ".patrolbot")
			path := filepath.Join(dir, "config.yaml")

			_, err = os.Stat(dir)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("error checking home config directory: %w", err)
			}

			exists := err == nil
			if exists {
				overwrite, err := promptOverwrite(cmd.Context(), dir)
				if err != nil {
					return err
				}

				if !overwrite {
					return nil
				}
			}

			err = copyTemplates(dir)
			if err != nil {
				return err
			}

			formData, err := promptConfig(cmd.Context())
			if err != nil {
				return err
			}

			err = writeConfig(formData, path)
			if err != nil {
				return err
			}

			cmd.Println("Wrote configuration file: " + path)

			return nil
		},
	}
}

func copyTemplates(dir string) error {
	//nolint: wrapcheck
	// error returned from WalkDir is wrapped once instead of wrapping each specific error
	err := fs.WalkDir(templateDirFS, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(dir, path)

		if entry.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		src, err := templateDirFS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close() //nolint:errcheck

		dest, err := os.Create(filepath.Clean(destPath))
		if err != nil {
			return err
		}
		defer dest.Close() //nolint:errcheck

		_, err = io.Copy(dest, src)

		return err
	})
	if err != nil {
		return fmt.Errorf("error copying report templates to home config directory: %w", err)
	}

	return nil
}

func writeConfig(formData formData, path string) error {
	tmpl, err := template.ParseFS(configTemplateFS, "config.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing config template: %w", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, formData)
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
