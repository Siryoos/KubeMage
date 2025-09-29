package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	
	"github.com/siryoos/kubemage/internal/llm"
)

type ModelSettings struct {
	Chat          string  `yaml:"chat"`
	Generation    string  `yaml:"generation"`
	Temperature   float64 `yaml:"temperature"`
	TopP          float64 `yaml:"top_p"`
	RepeatPenalty float64 `yaml:"repeat_penalty"`
}

type TruncationSettings struct {
	Message int `yaml:"message"`
	Logs    int `yaml:"logs"`
}

type IntelligenceSettings struct {
	Enabled             bool    `yaml:"enabled"`
	ContextCacheTTL     int     `yaml:"context_cache_ttl"`    // seconds
	ConfidenceThreshold float64 `yaml:"confidence_threshold"` // 0.0-1.0
	AutoOptimization    bool    `yaml:"auto_optimization"`
	LearningEnabled     bool    `yaml:"learning_enabled"`
	PlaybooksEnabled    bool    `yaml:"playbooks_enabled"`
	RiskAssessment      bool    `yaml:"risk_assessment"`
	QuickActionsEnabled bool    `yaml:"quick_actions_enabled"`
}

type PerformanceSettings struct {
	MaxConcurrent   int `yaml:"max_concurrent"`   // Max concurrent operations
	CommandTimeout  int `yaml:"command_timeout"`  // seconds
	ResponseTimeout int `yaml:"response_timeout"` // seconds
	RenderThrottle  int `yaml:"render_throttle"`  // milliseconds
	MemoryLimit     int `yaml:"memory_limit"`     // MB
}

type legacyPreferences struct {
	Theme string `yaml:"theme"`
}

type AppConfig struct {
	Models        ModelSettings        `yaml:"models"`
	NumCtx        int                  `yaml:"num_ctx"`
	KeepAlive     string               `yaml:"keep_alive"`
	Truncation    TruncationSettings   `yaml:"truncation"`
	Intelligence  IntelligenceSettings `yaml:"intelligence"`
	Performance   PerformanceSettings  `yaml:"performance"`
	Theme         string               `yaml:"theme"`
	HistoryLength int                  `yaml:"history_length"`
	OllamaHost    string               `yaml:"ollama_host,omitempty"`

	LegacyModel       string             `yaml:"model,omitempty"`
	LegacyTruncation  int                `yaml:"truncation_size,omitempty"`
	LegacyHistory     int                `yaml:"chat_history_length,omitempty"`
	LegacyPreferences *legacyPreferences `yaml:"preferences,omitempty"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		Models: ModelSettings{
			Chat:          "llama3.1:8b",
			Generation:    "llama3.1:13b",
			Temperature:   0.25,
			TopP:          0.9,
			RepeatPenalty: 1.1,
		},
		NumCtx:    12288, // Increased for intelligence features
		KeepAlive: "30m", // Longer for learning sessions
		Truncation: TruncationSettings{
			Message: 1200,
			Logs:    200,
		},
		Intelligence: IntelligenceSettings{
			Enabled:             true,
			ContextCacheTTL:     30,    // 30 seconds
			ConfidenceThreshold: 0.7,   // 70% confidence threshold
			AutoOptimization:    false, // User approval required
			LearningEnabled:     true,
			PlaybooksEnabled:    true,
			RiskAssessment:      true,
			QuickActionsEnabled: true,
		},
		Performance: PerformanceSettings{
			MaxConcurrent:   3,   // 3 concurrent operations
			CommandTimeout:  8,   // 8 seconds for commands
			ResponseTimeout: 120, // 2 minutes for LLM responses
			RenderThrottle:  40,  // 40ms render throttle
			MemoryLimit:     512, // 512MB memory limit
		},
		Theme:         "default",
		HistoryLength: 10,
		OllamaHost:    llm.DefaultOllamaEndpoint,
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
		if os.IsNotExist(err) {
			// Config file doesn't exist, return default config
			cfg := DefaultConfig()
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config.yaml: %w", err)
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
