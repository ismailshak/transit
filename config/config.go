package config

import (
	"github.com/spf13/viper"
)

type CoreConfig struct {
	ApiKey   string `mapstructure:"api_key"`
	Location string `mapstructure:"location"`
}

type Config struct {
	Core CoreConfig `mapstructure:"core"`
}

var (
	config Config
	vp     *viper.Viper
)

func GetConfig() *Config {
	return &config
}

func LoadConfig(path *string) (*Config, error) {
	vp = viper.New()

	if *path != "" {
		vp.SetConfigFile(*path)
	} else {
		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath("$HOME/.config/transit/")
	}

	err := vp.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = vp.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	return &config, nil
}
