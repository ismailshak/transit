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

// Options for the `dmv` section of a user config file
type SFConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

// Options for the `core` section of a user config file
type CoreConfig struct {
	Location      string `mapstructure:"location"`
	WatchInterval int    `mapstructure:"watch_interval"`
	Verbose       bool
}

type Config struct {
	Core CoreConfig `mapstructure:"core"`
	DMV  DmvConfig  `mapstructure:"dmv"`
	SF   DmvConfig  `mapstructure:"sf"`
}

var (
	config             *Config
	vp                 *viper.Viper = viper.New()
	configFileOverride string
)

// Returns the user's config singleton object
func GetConfig() *Config {
	if config == nil {
		config = &Config{}
	}

	return config
}

// Returns the config file that was used to load in user options
func GetConfigFileUsed() string {
	return vp.ConfigFileUsed()
}

// Returns the location of transit's config directory
func GetConfigDir() (string, error) {
	return getDefaultConfigDir()
}

// Gets a config value from user input. Nested fields are addressable
// by using a dot (.) as a delimiter e.g. `core.location`
func GetValue(key string) interface{} {
	return vp.Get(key)
}

// Set a config value from user input. Nested fields are addressable
// by using a dot (.) as a delimiter e.g. `core.location`
func SetValue(key, value string) error {
	vp.Set(key, value)
	err := vp.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write config file: %s", err)
	}

	verbose := config.Core.Verbose
	err = vp.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config after write: %s", err)
	}

	// This is a hack to ensure that the verbose flag is not overwritten by unmashalling
	// TODO: Find a better way to handle this
	config.Core.Verbose = verbose

	return nil
}

// Reads the config file and loads it's content
func LoadConfig() error {
	if configFileOverride != "" {
		vp.SetConfigFile(configFileOverride)
		setDefaults()
		return readConfig()
	}

	configDir, err := getDefaultConfigDir()
	if err != nil {
		return err
	}

	if !configFileExists(configDir) {
		err = utils.CreatePathIfNotFound(filepath.Join(configDir, "config.yml"))
		if err != nil {
			return err
		}
	}

	vp.SetConfigName("config")
	vp.AddConfigPath(configDir)

	setDefaults()

	return readConfig()
}

func SetCustomConfigPath(path string) {
	configFileOverride = path
}

func getDefaultConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "transit"), nil
}

func readConfig() error {
	err := vp.ReadInConfig()
	if err != nil {
		return err
	}

	err = vp.Unmarshal(&config)
	if err != nil {
		return err
	}

	return nil
}

func setDefaults() {
	vp.SetDefault("core.watch_interval", 10)
	vp.SetDefault("core.verbose", false)
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
