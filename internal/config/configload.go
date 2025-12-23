package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ErrFileNotFound indicates that no config file was found in the home directory, current directory, or at the --config
// flag path if given.
var ErrFileNotFound = errors.New("config.yaml file not found")

// Load sources config from files, environment variables, and command-line flags and returns the merged result.
// See patrolbot config --help for more on config sources.
func Load(flags *pflag.FlagSet) (Config, error) {
	viper.SetEnvPrefix("PATROLBOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
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

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}
