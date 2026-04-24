package main

import (
	"testing"
)

func TestLoadTGCConfig_Defaults(t *testing.T) {
	cfg := LoadTGCConfig()
	if cfg.API.BaseURL == "" {
		t.Error("base URL should have default")
	}
	if cfg.Models.Super == "" {
		t.Error("model should have default")
	}
}

func TestLoadTGCConfig_DefaultBaseURL(t *testing.T) {
	cfg := LoadTGCConfig()
	if cfg.API.BaseURL == "" {
		t.Error("base URL should not be empty")
	}
}

func TestLoadTGCConfig_DefaultModel(t *testing.T) {
	cfg := LoadTGCConfig()
	if cfg.Models.Super == "" {
		t.Error("model should not be empty")
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", "****"},
		{"short", "****"},
		{"sk-1234567890abcdef", "sk-1...cdef"},
	}
	for _, tt := range tests {
		got := maskKey(tt.in)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
