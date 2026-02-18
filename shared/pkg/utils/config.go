package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig[T any](path string) (*T, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file: %w", err)
	}

	var config T
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("Error parsing config file: %w", err)
	}

	return &config, nil
}
