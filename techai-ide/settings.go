// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE — github.com/flykimjiwon/TECHAI_CODE
// Forked from Hanimo Code: github.com/flykimjiwon/hanimo

package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GetSettings returns the current config for the Settings UI.
func (a *App) GetSettings() map[string]string {
	cfg := LoadTGCConfig()
	return map[string]string{
		"baseURL":  cfg.API.BaseURL,
		"apiKey":   maskKey(cfg.API.APIKey),
		"model":    cfg.Models.Super,
		"language": userLanguage,
	}
}

// SetLanguage changes the AI response language.
func (a *App) SetLanguage(lang string) {
	userLanguage = lang
	// Reinitialize chat with new system prompt
	a.chatMu.Lock()
	if a.chat != nil {
		a.chat.history[0].Content = systemPrompt()
	}
	a.chatMu.Unlock()
}

// SaveSettings writes updated settings to config.yaml.
func (a *App) SaveSettings(baseURL, apiKey, model string) error {
	cfg := LoadTGCConfig()

	if baseURL != "" {
		cfg.API.BaseURL = baseURL
	}
	if apiKey != "" && apiKey != maskKey(cfg.API.APIKey) {
		cfg.API.APIKey = apiKey
	}
	if model != "" {
		cfg.Models.Super = model
		cfg.Models.Dev = model
	}

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tgc")
	_ = os.MkdirAll(dir, 0755)

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0600); err != nil {
		return err
	}

	// Reinitialize chat engine with new config (thread-safe)
	a.chatMu.Lock()
	a.chat = newChatEngine(cfg, a)
	a.chatMu.Unlock()
	return nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
