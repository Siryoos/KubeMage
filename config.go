package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ModelSettings struct {
	Chat       string `yaml:"chat"`
	Generation string `yaml:"generation"`
}

type TruncationSettings struct {
	Message int `yaml:"message"`
}

type legacyPreferences struct {
	Theme string `yaml:"theme"`
}

type AppConfig struct {
	Models        ModelSettings      `yaml:"models"`
	NumCtx        int                `yaml:"num_ctx"`
	KeepAlive     string             `yaml:"keep_alive"`
	Truncation    TruncationSettings `yaml:"truncation"`
	Theme         string             `yaml:"theme"`
	HistoryLength int                `yaml:"history_length"`
	OllamaHost    string             `yaml:"ollama_host,omitempty"`

	LegacyModel       string             `yaml:"model,omitempty"`
	LegacyTruncation  int                `yaml:"truncation_size,omitempty"`
	LegacyHistory     int                `yaml:"chat_history_length,omitempty"`
	LegacyPreferences *legacyPreferences `yaml:"preferences,omitempty"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		Models: ModelSettings{
			Chat:       "llama3.1:8b",
			Generation: "llama3.1:13b",
		},
		NumCtx:        4096,
		KeepAlive:     "5m",
		Truncation:    TruncationSettings{Message: 1200},
		Theme:         "default",
		HistoryLength: 10,
		OllamaHost:    defaultOllamaEndpoint,
	}
}

func (cfg *AppConfig) applyDefaults() {
	defaults := DefaultConfig()

	if cfg.Models.Chat == "" {
		if cfg.LegacyModel != "" {
			cfg.Models.Chat = cfg.LegacyModel
		} else {
			cfg.Models.Chat = defaults.Models.Chat
		}
	}
	if cfg.Models.Generation == "" {
		cfg.Models.Generation = defaults.Models.Generation
	}
	if cfg.NumCtx == 0 {
		cfg.NumCtx = defaults.NumCtx
	}
	if strings.TrimSpace(cfg.KeepAlive) == "" {
		cfg.KeepAlive = defaults.KeepAlive
	}
	if cfg.Truncation.Message == 0 {
		if cfg.LegacyTruncation != 0 {
			cfg.Truncation.Message = cfg.LegacyTruncation
		} else {
			cfg.Truncation.Message = defaults.Truncation.Message
		}
	}
	if cfg.HistoryLength == 0 {
		if cfg.LegacyHistory != 0 {
			cfg.HistoryLength = cfg.LegacyHistory
		} else {
			cfg.HistoryLength = defaults.HistoryLength
		}
	}
	if strings.TrimSpace(cfg.Theme) == "" {
		if cfg.LegacyPreferences != nil && cfg.LegacyPreferences.Theme != "" {
			cfg.Theme = cfg.LegacyPreferences.Theme
		} else {
			cfg.Theme = defaults.Theme
		}
	}
	if strings.TrimSpace(cfg.OllamaHost) == "" {
		cfg.OllamaHost = defaults.OllamaHost
	}
}

func LoadConfig() (*AppConfig, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		cfg := DefaultConfig()
		return cfg, nil
	}
	defer f.Close()

	var cfg AppConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func SaveConfig(cfg *AppConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	cfg.LegacyModel = ""
	cfg.LegacyTruncation = 0
	cfg.LegacyHistory = 0
	cfg.LegacyPreferences = nil

	dir := filepath.Dir("config.yaml")
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	file, err := os.Create("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to create config.yaml: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func UpdateModelInConfig(scope, newModel string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	switch strings.ToLower(scope) {
	case "chat", "router", "default", "":
		cfg.Models.Chat = newModel
	case "generation", "strong", "diff":
		cfg.Models.Generation = newModel
	default:
		return fmt.Errorf("unknown model scope '%s'", scope)
	}

	if err := SaveConfig(cfg); err != nil {
		return err
	}
	SetActiveConfig(cfg)
	return nil
}

// Backward compatibility aliases

type Config = AppConfig

var activeConfig *AppConfig

func SetActiveConfig(cfg *AppConfig) {
	activeConfig = cfg
}

func ActiveConfig() *AppConfig {
	return activeConfig
}
