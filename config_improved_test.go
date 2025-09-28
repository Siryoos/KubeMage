// config_improved_test.go - Enhanced tests for config functionality
package main

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	
	if cfg.Models.Chat == "" {
		t.Error("DefaultConfig() should have a default chat model")
	}
	
	if cfg.Models.Generation == "" {
		t.Error("DefaultConfig() should have a default generation model")
	}
	
	if cfg.NumCtx == 0 {
		t.Error("DefaultConfig() should have a default context size")
	}
	
	if cfg.KeepAlive == "" {
		t.Error("DefaultConfig() should have a default keep alive time")
	}
	
	if cfg.Truncation.Message == 0 {
		t.Error("DefaultConfig() should have a default message truncation size")
	}
	
	if cfg.HistoryLength == 0 {
		t.Error("DefaultConfig() should have a default history length")
	}
	
	if cfg.Theme == "" {
		t.Error("DefaultConfig() should have a default theme")
	}
	
	if cfg.OllamaHost == "" {
		t.Error("DefaultConfig() should have a default Ollama host")
	}
}

func TestConfigApplyDefaults(t *testing.T) {
	cfg := &AppConfig{}
	cfg.applyDefaults()
	
	defaults := DefaultConfig()
	
	if cfg.Models.Chat != defaults.Models.Chat {
		t.Errorf("applyDefaults() chat model = %v, want %v", cfg.Models.Chat, defaults.Models.Chat)
	}
	
	if cfg.Models.Generation != defaults.Models.Generation {
		t.Errorf("applyDefaults() generation model = %v, want %v", cfg.Models.Generation, defaults.Models.Generation)
	}
	
	if cfg.NumCtx != defaults.NumCtx {
		t.Errorf("applyDefaults() num ctx = %v, want %v", cfg.NumCtx, defaults.NumCtx)
	}
	
	if cfg.KeepAlive != defaults.KeepAlive {
		t.Errorf("applyDefaults() keep alive = %v, want %v", cfg.KeepAlive, defaults.KeepAlive)
	}
	
	if cfg.Truncation.Message != defaults.Truncation.Message {
		t.Errorf("applyDefaults() message truncation = %v, want %v", cfg.Truncation.Message, defaults.Truncation.Message)
	}
	
	if cfg.HistoryLength != defaults.HistoryLength {
		t.Errorf("applyDefaults() history length = %v, want %v", cfg.HistoryLength, defaults.HistoryLength)
	}
	
	if cfg.Theme != defaults.Theme {
		t.Errorf("applyDefaults() theme = %v, want %v", cfg.Theme, defaults.Theme)
	}
	
	if cfg.OllamaHost != defaults.OllamaHost {
		t.Errorf("applyDefaults() ollama host = %v, want %v", cfg.OllamaHost, defaults.OllamaHost)
	}
}

func TestConfigLegacySupport(t *testing.T) {
	cfg := &AppConfig{
		LegacyModel:       "legacy-model",
		LegacyTruncation:  1000,
		LegacyHistory:     5,
		LegacyPreferences: &legacyPreferences{Theme: "legacy-theme"},
	}
	
	cfg.applyDefaults()
	
	if cfg.Models.Chat != "legacy-model" {
		t.Errorf("applyDefaults() should use legacy model, got %v", cfg.Models.Chat)
	}
	
	if cfg.Truncation.Message != 1000 {
		t.Errorf("applyDefaults() should use legacy truncation, got %v", cfg.Truncation.Message)
	}
	
	if cfg.HistoryLength != 5 {
		t.Errorf("applyDefaults() should use legacy history, got %v", cfg.HistoryLength)
	}
	
	if cfg.Theme != "legacy-theme" {
		t.Errorf("applyDefaults() should use legacy theme, got %v", cfg.Theme)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Test with a non-existent config file
	cfg, err := LoadConfig()
	
	if err != nil {
		t.Errorf("LoadConfig() with non-existent file should not return error, got %v", err)
	}
	
	if cfg == nil {
		t.Fatal("LoadConfig() should return default config for non-existent file")
	}
	
	// Should return default config
	defaults := DefaultConfig()
	if cfg.Models.Chat != defaults.Models.Chat {
		t.Errorf("LoadConfig() should return default config, got %v", cfg.Models.Chat)
	}
}

func TestSaveConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Models.Chat = "test-model"
	
	err := SaveConfig(cfg)
	if err != nil {
		t.Errorf("SaveConfig() failed: %v", err)
	}
	
	// Clean up
	os.Remove("config.yaml")
}

func TestSaveConfig_NilConfig(t *testing.T) {
	err := SaveConfig(nil)
	if err == nil {
		t.Error("SaveConfig() with nil config should return error")
	}
}

func TestUpdateModelInConfig(t *testing.T) {
	// Create a temporary config file
	cfg := DefaultConfig()
	cfg.Models.Chat = "original-chat"
	cfg.Models.Generation = "original-generation"
	
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove("config.yaml")
	
	// Test updating chat model
	err = UpdateModelInConfig("chat", "new-chat-model")
	if err != nil {
		t.Errorf("UpdateModelInConfig() failed: %v", err)
	}
	
	// Verify the update
	updatedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}
	
	if updatedCfg.Models.Chat != "new-chat-model" {
		t.Errorf("UpdateModelInConfig() chat model = %v, want %v", updatedCfg.Models.Chat, "new-chat-model")
	}
	
	// Test updating generation model
	err = UpdateModelInConfig("generation", "new-generation-model")
	if err != nil {
		t.Errorf("UpdateModelInConfig() failed: %v", err)
	}
	
	// Verify the update
	updatedCfg, err = LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}
	
	if updatedCfg.Models.Generation != "new-generation-model" {
		t.Errorf("UpdateModelInConfig() generation model = %v, want %v", updatedCfg.Models.Generation, "new-generation-model")
	}
}

func TestUpdateModelInConfig_InvalidScope(t *testing.T) {
	err := UpdateModelInConfig("invalid", "test-model")
	if err == nil {
		t.Error("UpdateModelInConfig() with invalid scope should return error")
	}
}

func TestActiveConfig(t *testing.T) {
	// Test initial state - may not be nil due to other tests
	cfg := ActiveConfig()
	
	// Set active config
	testCfg := DefaultConfig()
	SetActiveConfig(testCfg)
	
	// Test after setting
	cfg = ActiveConfig()
	if cfg != testCfg {
		t.Error("ActiveConfig() should return the set config")
	}
}
