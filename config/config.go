package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PollInterval string        `yaml:"poll_interval"`
	Workers      int           `yaml:"workers"`
	Accounts     []IMAPAccount `yaml:"accounts"`
}
type IMAPAccount struct {
	Name             string `yaml:"name"`
	Host             string `yaml:"host"`
	TLS              bool   `yaml:"tls"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Inbox            string `yaml:"inbox"`
	SpamFolder       string `yaml:"spam_folder"`
	SearchUnseenOnly bool   `yaml:"search_unseen_only"`
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}
