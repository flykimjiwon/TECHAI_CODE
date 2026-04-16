package tools

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// Secret patterns to detect before writing files or committing.
var secretPatterns = []struct {
	re   *regexp.Regexp
	desc string
}{
	{regexp.MustCompile(`(?i)(sk-[a-zA-Z0-9]{20,})`), "OpenAI API key"},
	{regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`), "AWS Access Key"},
	{regexp.MustCompile(`-----BEGIN (RSA |EC |DSA )?PRIVATE KEY-----`), "Private key"},
	{regexp.MustCompile(`(?i)(ghp_[a-zA-Z0-9]{36})`), "GitHub token"},
	{regexp.MustCompile(`(?i)(gho_[a-zA-Z0-9]{36})`), "GitHub OAuth token"},
	{regexp.MustCompile(`(?i)(glpat-[a-zA-Z0-9\-]{20,})`), "GitLab token"},
	{regexp.MustCompile(`(?i)password\s*[:=]\s*["'][^"']{8,}["']`), "Hardcoded password"},
	{regexp.MustCompile(`(?i)(xox[bporas]-[a-zA-Z0-9\-]{10,})`), "Slack token"},
	{regexp.MustCompile(`(?i)(eyJ[a-zA-Z0-9_-]{20,}\.[a-zA-Z0-9_-]{20,}\.)`), "JWT token"},
	{regexp.MustCompile(`(?i)(tg-[a-zA-Z0-9]{20,})`), "TechAI API key"},
}

// Sensitive file extensions that should not be written by AI.
var sensitiveFiles = map[string]bool{
	".env":             true,
	".env.local":       true,
	".env.production":  true,
	".pem":             true,
	".key":             true,
	".p12":             true,
	".pfx":             true,
	".jks":             true,
	"credentials.json": true,
	"service-account.json": true,
}

// CheckSecrets scans content for potential secrets.
// Returns a warning string if secrets are found, empty string if clean.
func CheckSecrets(content string) string {
	var found []string
	for _, p := range secretPatterns {
		if p.re.MatchString(content) {
			found = append(found, p.desc)
			config.DebugLog("[SECRET-WARN] detected: %s", p.desc)
		}
	}
	if len(found) == 0 {
		return ""
	}
	return "⚠ Potential secrets detected: " + strings.Join(found, ", ")
}

// CheckSensitiveFile returns a warning if the file path is a sensitive file type.
func CheckSensitiveFile(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	if sensitiveFiles[base] || sensitiveFiles[ext] {
		config.DebugLog("[SECRET-WARN] sensitive file: %s", path)
		return "⚠ Sensitive file: " + base
	}
	return ""
}
