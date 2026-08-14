// Package config provides an interface to get and set user configuration options.
//
// This package should be used by all packages when a certain value can be overridden by
// a user's config file. It should also not import any of the other transit packages
// to avoid import cycles (except utils).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ismailshak/transit/internal/utils"
	"github.com/spf13/viper"
)

// Options for the `dmv` section of a user config file
type DmvConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

// Options for the `sf` section of a user config file
type SFConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

// Options for the `core` section of a user config file
type CoreConfig struct {
	Location      string `mapstructure:"location"`
	WatchInterval int    `mapstructure:"watch_interval"`
}

type Config struct {
	Core CoreConfig `mapstructure:"core"`
	DMV  DmvConfig  `mapstructure:"dmv"`
	SF   SFConfig   `mapstructure:"sf"`

	// The file these values were decoded from. Kept so that Get and Set can
	// address keys by a runtime string, which a struct can't do.
	vp *viper.Viper
}

// Load reads the config file and decodes it into a Config. An override path is
// used as-is when non-empty, otherwise the default config directory is used and
// an empty config file is created there if none exists yet.
func Load(override string) (*Config, error) {
	vp := viper.New()
	c := &Config{vp: vp}

	c.setDefaults()

	if override != "" {
		vp.SetConfigFile(override)
		if err := c.read(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
		}

		return c, nil
	}

	configDir, err := getDefaultConfigDir()
	if err != nil {
		return nil, err
	}

	// TODO: We should also create the override if it doesn't exist
	if !configFileExists(configDir) {
		err = utils.CreatePathIfNotFound(filepath.Join(configDir, "config.yml"))
		if err != nil {
			return nil, err
		}
	}

	vp.SetConfigName("config")
	vp.AddConfigPath(configDir)

	if err := c.read(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return c, nil
}

// Get returns a config value looked up by a key from user input, or nil if the
// key isn't set. Nested fields are addressable by using a dot (.) as a
// delimiter e.g. `core.location`
func (c *Config) Get(key string) any {
	return c.vp.Get(key)
}

// Set writes a config value from user input back to the file and re-decodes it,
// so the file and this Config agree once it returns. Nested fields are
// addressable by using a dot (.) as a delimiter e.g. `core.location`
func (c *Config) Set(key, value string) error {
	c.vp.Set(key, value)
	if err := c.vp.WriteConfig(); err != nil {
		return fmt.Errorf("write config %s: %w", c.vp.ConfigFileUsed(), err)
	}

	if err := c.vp.Unmarshal(c); err != nil {
		return fmt.Errorf("decode config after write: %w", err)
	}

	return nil
}

// FileUsed returns the path to the config file these values were loaded from
func (c *Config) FileUsed() string {
	return c.vp.ConfigFileUsed()
}

func (c *Config) setDefaults() {
	c.vp.SetDefault("core.watch_interval", 10)
}

func (c *Config) read() error {
	err := c.vp.ReadInConfig()
	if err != nil {
		return err
	}

	err = c.vp.Unmarshal(c)
	if err != nil {
		return err
	}

	return nil
}

// GetConfigDir returns the location of transit's config directory
func GetConfigDir() (string, error) {
	return getDefaultConfigDir()
}

func getDefaultConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "transit"), nil
}

func configFileExists(baseDir string) bool {
	// This is the order of precedence for config files
	allowedFileTypes := []string{".yml", ".yaml", ".json", ".toml", ".ini"}

	for _, ft := range allowedFileTypes {
		if utils.FileExists(filepath.Join(baseDir, "config"+ft)) {
			return true
		}
	}

	return false
}
