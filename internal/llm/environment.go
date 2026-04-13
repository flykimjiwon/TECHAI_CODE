package llm

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// probe defines a tool to check: binary name + version flag.
type probe struct {
	Name    string
	Bin     string
	Version string // flag to get version (e.g. "--version")
}

// allProbes is the full list of runtimes/tools to detect on startup.
var allProbes = []probe{
	// JavaScript / TypeScript
	{"node", "node", "--version"},
	{"npm", "npm", "--version"},
	{"npx", "npx", "--version"},
	{"pnpm", "pnpm", "--version"},
	{"yarn", "yarn", "--version"},
	{"bun", "bun", "--version"},
	{"deno", "deno", "--version"},

	// Python
	{"python3", "python3", "--version"},
	{"python", "python", "--version"},
	{"pip", "pip3", "--version"},
	{"uv", "uv", "--version"},
	{"poetry", "poetry", "--version"},

	// Go
	{"go", "go", "version"},

	// Rust
	{"rustc", "rustc", "--version"},
	{"cargo", "cargo", "--version"},

	// Java / JVM
	{"java", "java", "-version"},
	{"javac", "javac", "-version"},
	{"mvn", "mvn", "--version"},
	{"gradle", "gradle", "--version"},
	{"kotlin", "kotlin", "-version"},

	// Ruby
	{"ruby", "ruby", "--version"},
	{"gem", "gem", "--version"},
	{"bundle", "bundle", "--version"},

	// PHP
	{"php", "php", "--version"},
	{"composer", "composer", "--version"},

	// Swift / Dart / Flutter
	{"swift", "swift", "--version"},
	{"dart", "dart", "--version"},
	{"flutter", "flutter", "--version"},

	// .NET
	{"dotnet", "dotnet", "--version"},

	// Others
	{"lua", "lua", "-v"},
	{"perl", "perl", "--version"},
	{"R", "R", "--version"},
	{"zig", "zig", "version"},
	{"elixir", "elixir", "--version"},

	// Build tools
	{"make", "make", "--version"},
	{"cmake", "cmake", "--version"},

	// Containers & Git
	{"docker", "docker", "--version"},
	{"git", "git", "--version"},

	// Network
	{"curl", "curl", "--version"},

	// Cloud CLIs
	{"aws", "aws", "--version"},
	{"gcloud", "gcloud", "--version"},
	{"kubectl", "kubectl", "version --client --short"},
	{"terraform", "terraform", "--version"},
}

// ProbeResult holds the detection result for one tool.
type ProbeResult struct {
	Name      string
	Available bool
	Version   string // short version string or ""
}

// EnvProbeResults is the cached probe from startup.
var EnvProbeResults []ProbeResult

// ProbeEnvironment checks which tools are installed and caches results.
// All probes run concurrently with a 2s per-probe timeout, so total
// wall time is ~max(single probe) instead of sum(all probes).
// This is critical on Windows where CreateProcess is much slower.
func ProbeEnvironment() []ProbeResult {
	type indexed struct {
		idx    int
		result ProbeResult
	}

	var wg sync.WaitGroup
	ch := make(chan indexed, len(allProbes))

	for i, p := range allProbes {
		wg.Add(1)
		go func(idx int, p probe) {
			defer wg.Done()
			path, err := exec.LookPath(p.Bin)
			if err != nil || path == "" {
				ch <- indexed{idx, ProbeResult{Name: p.Name, Available: false}}
				return
			}
			ver := getVersion(p.Bin, p.Version)
			ch <- indexed{idx, ProbeResult{Name: p.Name, Available: true, Version: ver}}
		}(i, p)
	}

	go func() { wg.Wait(); close(ch) }()

	// Collect results preserving original order
	tmp := make([]ProbeResult, len(allProbes))
	for r := range ch {
		tmp[r.idx] = r.result
	}

	// Deduplicate: python3 vs python — keep whichever was found first
	seen := map[string]bool{}
	results := make([]ProbeResult, 0, len(allProbes))
	for _, r := range tmp {
		if r.Name == "python" && seen["python3"] {
			continue
		}
		if r.Available {
			seen[r.Name] = true
		}
		results = append(results, r)
	}

	EnvProbeResults = results
	return results
}

// getVersion runs "bin flag" with a 2s timeout and extracts a short
// version string. The timeout prevents slow tools (gradle, flutter)
// from blocking startup.
func getVersion(bin, flag string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	args := strings.Fields(flag)
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	line = strings.TrimPrefix(line, "v")
	if len([]rune(line)) > 40 {
		line = string([]rune(line)[:40])
	}
	return line
}

// FormatEnvironmentContext returns a compact system-prompt block showing
// which tools are available. The LLM reads this to avoid calling
// tools that aren't installed.
func FormatEnvironmentContext(results []ProbeResult) string {
	var available []string
	var missing []string

	for _, r := range results {
		if r.Available {
			label := r.Name
			if r.Version != "" {
				label += " " + r.Version
			}
			available = append(available, label)
		} else {
			missing = append(missing, r.Name)
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n## Environment: %s/%s\n", runtime.GOOS, runtime.GOARCH))

	if len(available) > 0 {
		b.WriteString("Installed: " + strings.Join(available, " | ") + "\n")
	}
	if len(missing) > 0 {
		b.WriteString("Not installed: " + strings.Join(missing, " | ") + "\n")
	}
	b.WriteString("미설치 도구는 사용하지 마세요. 필요하면 사용자에게 설치 방법을 안내하세요.\n")

	return b.String()
}
