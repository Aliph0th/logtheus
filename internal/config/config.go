package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Env    string `mapstructure:"ENV"`
	Server struct {
		Port int `mapstructure:"PORT"`
	} `mapstructure:",squash"`
	DB struct {
		Host     string `mapstructure:"POSTGRES_HOST"`
		Port     int    `mapstructure:"POSTGRES_PORT"`
		Name     string `mapstructure:"POSTGRES_DB"`
		User     string `mapstructure:"POSTGRES_USER"`
		Password string `mapstructure:"POSTGRES_PASSWORD"`
	} `mapstructure:",squash"`

	JWT struct {
		AccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
		RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
		Issuer        string `mapstructure:"JWT_ISSUER"`
	} `mapstructure:",squash"`

	Redis struct {
		Password string `mapstructure:"REDIS_PASSWORD"`
		Host     string `mapstructure:"REDIS_HOST"`
		Port     int    `mapstructure:"REDIS_PORT"`
		Database int    `mapstructure:"REDIS_DATABASE"`
	} `mapstructure:",squash"`
}

func LoadConfig(path string) (*AppConfig, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file: %w", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("Error parsing config file %w", err)
	}

	return &config, nil
}
