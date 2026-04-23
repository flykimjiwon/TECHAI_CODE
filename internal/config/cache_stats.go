package config

import "sync/atomic"

// CacheStats tracks prompt cache performance across the session.
// Lives in config (not llm) to avoid circular imports.
type CacheStats struct {
	Requests     atomic.Int64
	CachedTokens atomic.Int64
	TotalTokens  atomic.Int64
}

// GlobalCacheStats accumulates cache metrics for the whole process.
var GlobalCacheStats CacheStats

// RecordCacheHit logs a request's cache performance.
func (cs *CacheStats) RecordCacheHit(cached, total int) {
	cs.Requests.Add(1)
	cs.CachedTokens.Add(int64(cached))
	cs.TotalTokens.Add(int64(total))
}

// CacheHitRate returns the fraction of input tokens served from cache.
func (cs *CacheStats) CacheHitRate() float64 {
	total := cs.TotalTokens.Load()
	if total == 0 {
		return 0
	}
	return float64(cs.CachedTokens.Load()) / float64(total)
}

// CacheSummary returns a one-line human-readable cache summary.
func CacheSummary() string {
	cs := &GlobalCacheStats
	reqs := cs.Requests.Load()
	cached := cs.CachedTokens.Load()
	total := cs.TotalTokens.Load()
	if reqs == 0 {
		return "캐시: 아직 요청 없음"
	}
	rate := int64(0)
	if total > 0 {
		rate = cached * 100 / total
	}
	return "캐시: " + i64str(cached) + "/" + i64str(total) + " 토큰 (" + i64str(rate) + "%)"
}

func i64str(n int64) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
