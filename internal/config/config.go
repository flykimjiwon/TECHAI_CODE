package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
}

type ModelsConfig struct {
	Super string `yaml:"super"`
	Dev   string `yaml:"dev"`
}

type Config struct {
	API    APIConfig    `yaml:"api"`
	Models ModelsConfig `yaml:"models"`
}

func DefaultConfig() Config {
	return Config{
		API: APIConfig{
			BaseURL: "https://api.novita.ai/openai",
			APIKey:  "",
		},
		Models: ModelsConfig{
			Super: "openai/gpt-oss-120b",
			Dev:   "qwen/qwen-2.5-coder-32b-instruct",
		},
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tgc")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func Load() (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("config parse error: %w", err)
		}
	}

	if v := os.Getenv("TGC_API_BASE_URL"); v != "" {
		cfg.API.BaseURL = v
	}
	if v := os.Getenv("TGC_API_KEY"); v != "" {
		cfg.API.APIKey = v
	}
	if v := os.Getenv("TGC_MODEL_SUPER"); v != "" {
		cfg.Models.Super = v
	}
	if v := os.Getenv("TGC_MODEL_DEV"); v != "" {
		cfg.Models.Dev = v
	}

	return cfg, nil
}

func Save(cfg Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func NeedsSetup() bool {
	cfg, err := Load()
	if err != nil {
		return true
	}
	return cfg.API.APIKey == "" && os.Getenv("TGC_API_KEY") == ""
}

func RunSetupWizard() (Config, error) {
	cfg := DefaultConfig()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n  택가이코드 설정")

	fmt.Print("  API Base URL [https://api.novita.ai/openai]: ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.API.BaseURL = strings.TrimSpace(input)
	}

	fmt.Print("  API Key: ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.API.APIKey = strings.TrimSpace(input)
	}

	if err := Save(cfg); err != nil {
		return cfg, err
	}

	fmt.Printf("\n  저장됨: %s\n\n", ConfigPath())
	return cfg, nil
}
