package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type DmvConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

type CoreConfig struct {
	Location string `mapstructure:"location"`
}

type Config struct {
	Core CoreConfig `mapstructure:"core"`
	Dmv  DmvConfig  `mapstructure:"dmv"`
}

var (
	config Config
	vp     *viper.Viper = viper.New()
)

// Entry point for `config get`
func ExecuteGet(key string) {
	result := vp.Get(key)

	if result == nil {
		fmt.Printf("No config property found matching '%s'\n", key)
		return
	}

	fmt.Println(result)
}

// Entry point for `config set`
func ExecuteSet(arg string) {
	key, value, valid := parseSetArg(arg)
	if !valid {
		fmt.Printf("Could not parse '%s'. Make sure it's in the format <key>=<value>\n", key)
		return
	}

	valid = validateKey(key, value)
	if !valid {
		return
	}

	vp.Set(key, value)
	vp.WriteConfig()
	fmt.Printf("'%s' has been set to '%s'\n", key, value)
}

// Entry point for `config path`
func ExecutePath() {
	fmt.Println(getConfigPath())
}

func GetConfig() *Config {
	return &config
}

// This should be called before any command gets parsed & executed
func LoadConfig(path *string) (*Config, error) {
	if *path != "" {
		vp.SetConfigFile(*path)
	} else {
		createConfigIfNotFound(os.Getenv("HOME") + "/.config/transit/config.yml")
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

func parseSetArg(arg string) (string, string, bool) {
	parts := strings.Split(arg, "=")

	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

var validLocations = map[string]bool{
	"dmv": true,
}

func validateKey(key, value string) bool {
	if key == "core.location" {
		valid := validateLocation(value)
		if !valid {
			fmt.Printf("'%s' is not a valid location\n", value)
			fmt.Printf("Valid locations: %s\n", reflect.ValueOf(validLocations).MapKeys())
			return false
		}
	}

	return true
}

func validateLocation(location string) bool {
	_, exists := validLocations[location]
	return exists
}

func getConfigPath() string {
	return vp.ConfigFileUsed()
}

func createConfigIfNotFound(configPath string) {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Config not found. Creating a config file")
		os.WriteFile(configPath, []byte{}, fs.FileMode(os.O_CREATE))
	}
}
