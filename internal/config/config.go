package config

type DmvConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

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
