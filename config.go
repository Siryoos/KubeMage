package main

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Model             string `yaml:"model"`
	MaxTokens         int    `yaml:"max_tokens"`
	TruncationSize    int    `yaml:"truncation_size"`
	ChatHistoryLength int    `yaml:"chat_history_length"`
}

func LoadConfig() (*Config, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
