package llm

import "sync"

// Prompt caching strategy for 택가이코드.
//
// Policed-network environment: single OpenAI-compatible endpoint.
// Auto prefix caching (50% discount) is enabled via IncludeUsage.
//
// Strategy:
//   1. Keep system prompt prefix STABLE across turns.
//   2. Track the stable/dynamic boundary for future explicit cache control.
//   3. config.GlobalCacheStats tracks hits from IncludeUsage responses.

// CacheBreakpoint tracks where the "stable" part of the system prompt
// ends. Content before this offset can be marked as cacheable when
// talking to providers that support explicit cache control.
type CacheBreakpoint struct {
	mu            sync.RWMutex
	stablePrefix  string // core + mode body + askuser (invariant)
	dynamicSuffix string // project ctx + knowledge TOC + skills TOC
}

// GlobalBreakpoint tracks the system prompt cache boundary.
var GlobalBreakpoint CacheBreakpoint

// SetBreakpoint records the stable/dynamic split of the system prompt.
func (cb *CacheBreakpoint) SetBreakpoint(stable, dynamic string) {
	cb.mu.Lock()
	cb.stablePrefix = stable
	cb.dynamicSuffix = dynamic
	cb.mu.Unlock()
}

// StablePrefix returns the cacheable prefix of the system prompt.
func (cb *CacheBreakpoint) StablePrefix() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.stablePrefix
}

// Full returns the complete system prompt (stable + dynamic).
func (cb *CacheBreakpoint) Full() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.stablePrefix + cb.dynamicSuffix
}
