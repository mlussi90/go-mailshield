package main

import (
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

func loadConfig(path string) Config {
	f, err := os.Open(path)
	must(err)
	defer f.Close()

	var cfg Config
	must(yaml.NewDecoder(f).Decode(&cfg))
	return cfg
}
