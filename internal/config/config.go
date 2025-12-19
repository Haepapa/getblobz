package config

import (
	"strings"

	"github.com/spf13/viper"
)


type Config struct {
	ConnectionString string `mapstructure:"connection_string"`
	ContainerName	 string `mapstructure:"container_name"`
	LocalPath        string `mapstructure:"local_path"`
	Workers		 	 int    `mapstructure:"workers"`
}

func LoadConfig() (*Config, error) {
	// set defaults
	viper.SetDefault("workers", 5)
	viper.SetDefault("local_path", "./downloads")

	// look for a file named "getblobz.yaml"
	viper.SetConfigName("getblobz")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// enable env variables
	viper.SetEnvPrefix("GETBLOBZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// is a config files is found, read it in
	_ = viper.ReadInConfig()

	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}