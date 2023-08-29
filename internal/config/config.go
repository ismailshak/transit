// Package config provides an interface to get and set user configuration options.
//
// This package should be used by all packages when a certain value can be overridden by
// a user's config file. It should also not import any of the other transit packages
// to avoid import cycles.
package config

// Options for the `dmv` section of a user config file
type DmvConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

// Options for the `core` section of a user config file
type CoreConfig struct {
	Verbose       bool
	Location      string `mapstructure:"location"`
	WatchInterval int    `mapstructure:"watch_interval"`
}

type Config struct {
	Core CoreConfig `mapstructure:"core"`
	Dmv  DmvConfig  `mapstructure:"dmv"`
}

var config Config

func GetConfig() *Config {
	return &config
}

func SetConfig(newConfig *Config) {
	config = *newConfig
}
