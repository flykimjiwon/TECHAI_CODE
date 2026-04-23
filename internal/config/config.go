package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// AppVersion is set by main.go from the build-time ldflags version string.
// UI components read this to display the current version.
var AppVersion = "dev"

type APIConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
}

type ModelsConfig struct {
	Super string `yaml:"super"`
	Dev   string `yaml:"dev"`
}

type MultiConfig struct {
	Enabled  bool   `yaml:"enabled"`  // multi-agent on/off
	Strategy string `yaml:"strategy"` // "auto", "review", "consensus", "scan"
}

type MCPServer struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"` // "stdio" or "sse"
	Command   string            `yaml:"command"`   // for stdio: command to run
	Args      []string          `yaml:"args"`      // for stdio: command args
	URL       string            `yaml:"url"`       // for sse: endpoint URL
	Env       map[string]string `yaml:"env"`       // optional env vars
}

type MCPConfig struct {
	Servers []MCPServer `yaml:"servers"`
}

type Config struct {
	API    APIConfig    `yaml:"api"`
	Models ModelsConfig `yaml:"models"`
	Multi  MultiConfig  `yaml:"multi"`
	MCP    MCPConfig    `yaml:"mcp"`
}

// Build-time overridable defaults (set via -ldflags)
var (
	DefaultBaseURL  = "https://api.novita.ai/openai"
	DefaultModel    = "qwen/qwen3-coder-30b-a3b-instruct"
	DefaultDevModel = "qwen/qwen3-coder-30b-a3b-instruct"
	ConfigDirName   = ".tgc"
	DebugMode       = "false" // set to "true" via ldflags in build-debug
)

// IsDebug returns true when the binary was built with debug mode enabled.
func IsDebug() bool { return DebugMode == "true" }

var (
	debugFile *os.File
	debugMu   sync.Mutex
)

// InitDebugLog opens the debug log file in the config directory.
func InitDebugLog() {
	if !IsDebug() {
		return
	}
	dir := ConfigDir()
	_ = os.MkdirAll(dir, 0755)
	f, err := os.Create(filepath.Join(dir, "debug.log"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] failed to create debug.log: %v\n", err)
		return
	}
	debugFile = f
	DebugLog("========================================")
	DebugLog("=== TECHAI DEBUG MODE ===")
	DebugLog("========================================")
	DebugLog("[SYS] OS=%s | ARCH=%s | GoVersion=%s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	hostname, _ := os.Hostname()
	DebugLog("[SYS] Hostname=%s", hostname)
	home, _ := os.UserHomeDir()
	DebugLog("[SYS] HomeDir=%s", home)
	cwd, _ := os.Getwd()
	DebugLog("[SYS] CWD=%s", cwd)
	DebugLog("[SYS] ConfigDir=%s", dir)
	DebugLog("[SYS] PID=%d", os.Getpid())
	// Dump all TGC_ env vars
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "TGC_") || strings.HasPrefix(e, "HTTP") || strings.HasPrefix(e, "NO_PROXY") || strings.HasPrefix(e, "https_proxy") || strings.HasPrefix(e, "http_proxy") || strings.HasPrefix(e, "no_proxy") {
			DebugLog("[ENV] %s", e)
		}
	}
	DebugLog("========================================")
}

// CloseDebugLog flushes and closes the debug log file.
func CloseDebugLog() {
	debugMu.Lock()
	defer debugMu.Unlock()
	if debugFile != nil {
		debugFile.Close()
		debugFile = nil
	}
}

// DebugLog writes a timestamped line to the debug log file.
// No-op when debug mode is disabled or log file is not open.
func DebugLog(format string, args ...interface{}) {
	if !IsDebug() || debugFile == nil {
		return
	}
	debugMu.Lock()
	defer debugMu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(debugFile, "[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
}

// DebugLogPath returns the path to the debug log file.
func DebugLogPath() string {
	return filepath.Join(ConfigDir(), "debug.log")
}

func DefaultConfig() Config {
	return Config{
		API: APIConfig{
			BaseURL: DefaultBaseURL,
			APIKey:  "",
		},
		Models: ModelsConfig{
			Super: DefaultModel,
			Dev:   DefaultDevModel,
		},
		Multi: MultiConfig{
			Enabled:  false,
			Strategy: "auto",
		},
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ConfigDirName)
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
	// ── SUSPENDED: TGC_MULTI env var ignored until sub-agent re-open ──
	// if v := os.Getenv("TGC_MULTI"); v != "" { ... }
	cfg.Multi.Enabled = false

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

	fmt.Printf("  API Base URL [%s]: ", DefaultBaseURL)
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
