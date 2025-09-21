package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Engine  Engine        `yaml:"engine"`
		Network NetworkConfig `yaml:"network"`
		Logging LoggingConfig `yaml:"logging"`
	}

	Engine struct {
		Type string `yaml:"type" default:"in_memory"`
	}

	NetworkConfig struct {
		Address        string        `yaml:"address" default:"127.0.0.1:8080"`
		MaxConnections int           `yaml:"max_connections" default:"100"`
		MaxMessageSize int           `yaml:"max_message_size" default:"4KB"`
		IdleTimeout    time.Duration `yaml:"idle_timeout" default:"5m"`
	}
	LoggingConfig struct {
		Level  string `yaml:"level" default:"info"`
		Output string `yaml:"output" default:"stdout"`
	}
)

func New(configPath string) (*Config, error) {
	var cfg Config

	// read config
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &cfg, nil
}
