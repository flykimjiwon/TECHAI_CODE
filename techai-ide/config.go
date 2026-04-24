package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TGCConfig mirrors .tgc/config.yaml so we share settings with the TUI.
type TGCConfig struct {
	API struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
	} `yaml:"api"`
	Models struct {
		Super string `yaml:"super"`
		Dev   string `yaml:"dev"`
	} `yaml:"models"`
}

const (
	defaultBaseURL = "https://techai-web-prod.shinhan.com/v1"
	defaultModel   = "Qwen3-Coder-30B"
)

// LoadTGCConfig reads the shared .tgc/config.yaml (or .tgc-onprem/).
func LoadTGCConfig() TGCConfig {
	cfg := TGCConfig{}
	cfg.API.BaseURL = defaultBaseURL
	cfg.Models.Super = defaultModel
	cfg.Models.Dev = defaultModel

	home, _ := os.UserHomeDir()

	// Try config dirs in priority order
	dirs := []string{".tgc-onprem", ".tgc"}
	for _, dir := range dirs {
		path := filepath.Join(home, dir, "config.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		_ = yaml.Unmarshal(data, &cfg)
		break
	}

	// Env overrides
	if v := os.Getenv("TGC_API_BASE_URL"); v != "" {
		cfg.API.BaseURL = v
	}
	if v := os.Getenv("TGC_API_KEY"); v != "" {
		cfg.API.APIKey = v
	}
	if v := os.Getenv("TGC_MODEL_SUPER"); v != "" {
		cfg.Models.Super = v
	}

	return cfg
}
