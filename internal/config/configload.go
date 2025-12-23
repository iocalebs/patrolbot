package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Load sources config from files, environment variables, and command-line flags and returns the merged result.
// See patrolbot config --help for more on config sources.
func Load(cfgFile string, flags *pflag.FlagSet) (Config, error) {
	viper.SetEnvPrefix("PATROLBOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

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

	err := viper.ReadInConfig()
	if err != nil {
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
