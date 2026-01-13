// Package config provides functionality to load config from YAML files,
// environment variables, and command-line arguments.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/iocalebs/patrolbot/internal/cli/clierr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// ErrFileNotFound indicates that no config file was found in the home directory, current directory, or at the --config
	// flag path if given.
	ErrFileNotFound = errors.New("config.yaml file not found")

	// ErrWikiNotSet indicates that no wiki was selected via flag or env variable, and no default wiki is configured.
	ErrWikiNotSet = errors.New("wiki not set")

	// ErrWikiNotFound indicates that a wiki was selected that does not exist in the map of wikis.
	ErrWikiNotFound = errors.New("wiki not found")

	// ErrWikiInvalid indicates that the configuration of the selected wiki is invalid.
	ErrWikiInvalid = errors.New("invalid wiki configuration")
)

// Load sources config from files, environment variables, and command-line flags and returns the merged result.
// See patrolbot config --help for more on config sources.
func Load(flags *pflag.FlagSet) (Config, error) {
	viper.SetEnvPrefix("PATROLBOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	cfgFile, err := flags.GetString("config")
	if err != nil {
		return Config{}, fmt.Errorf("error reading config flag: %w", err)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("failed searching for config in home directory: %w", err)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.patrolbot")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	err = viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return Config{}, ErrFileNotFound
		}

		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	err = viper.BindPFlags(flags)
	if err != nil {
		cobra.CheckErr(err)
	}

	var cfg Config

	err = viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}

// CurrentWiki validates and returns the [Wiki] configuration for the selected wiki.
func (c Config) CurrentWiki() (Wiki, error) {
	if c.Wiki == "" {
		return Wiki{}, ErrWikiNotSet
	}

	wiki, ok := c.Wikis[c.Wiki]
	if !ok {
		return Wiki{}, fmt.Errorf("%w: %s", ErrWikiNotFound, c.Wiki)
	}

	errs := []string{}

	if wiki.Site.URL == "" {
		errs = append(errs, ".site.url not set")
	}

	if wiki.Auth.Username == "" {
		errs = append(errs, ".auth.username not set")
	}

	if wiki.Auth.Password == "" {
		errs = append(errs, ".auth.password not set")
	}

	if wiki.Reports.TemplateDir == "" {
		errs = append(errs, ".reports.templateDir not set")
	}

	if len(errs) > 0 {
		msg := fmt.Sprintf("invalid wiki configuration for %q", c.Wiki)
		return Wiki{}, clierr.NewMultiError(msg, errs)
	}

	return wiki, nil
}
