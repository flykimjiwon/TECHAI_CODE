package session

import (
	"fmt"
)

// usageSchema adds the usage_log table if it doesn't exist.
// Called lazily on first LogUsage call.
const usageSchema = `
CREATE TABLE IF NOT EXISTS usage_log (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT    NOT NULL,
	provider   TEXT    NOT NULL,
	model      TEXT    NOT NULL,
	tokens_in  INTEGER NOT NULL,
	tokens_out INTEGER NOT NULL,
	cost_usd   REAL    NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);
CREATE INDEX IF NOT EXISTS idx_usage_session ON usage_log(session_id);
`

// UsageSummary holds aggregated token/cost totals.
type UsageSummary struct {
	TokensIn  int64
	TokensOut int64
	CostUSD   float64
}

// pricing per 1M tokens (input, output) in USD.
// 폐쇄망 환경에서도 내부 서버 비용 추적용으로 사용.
var pricing = map[string][2]float64{
	// Novita.ai (default endpoint)
	"gpt-oss-120b":     {0.90, 0.90},
	"qwen3-coder-30b":  {0.30, 0.30},
	// Onprem models (internal billing)
	"shinhan-70b":      {0.00, 0.00},
	// Common cloud models (reference)
	"gpt-4o":           {2.50, 10.00},
	"gpt-4o-mini":      {0.15, 0.60},
	"gpt-4.1":          {2.00, 8.00},
	"gpt-4.1-mini":     {0.40, 1.60},
}

// estimateCost calculates the cost in USD for a given model and token counts.
func estimateCost(model string, tokensIn, tokensOut int) float64 {
	p, ok := pricing[model]
	if !ok {
		return 0
	}
	return (float64(tokensIn)*p[0] + float64(tokensOut)*p[1]) / 1_000_000
}

// EnsureUsageTable creates the usage_log table if missing.
func (s *Store) EnsureUsageTable() error {
	_, err := s.db.Exec(usageSchema)
	return err
}

// LogUsage records a usage entry for a session.
func (s *Store) LogUsage(sessionID, provider, model string, tokensIn, tokensOut int) error {
	// Ensure table exists (idempotent)
	if err := s.EnsureUsageTable(); err != nil {
		return fmt.Errorf("ensure usage table: %w", err)
	}
	cost := estimateCost(model, tokensIn, tokensOut)
	_, err := s.db.Exec(
		`INSERT INTO usage_log (session_id, provider, model, tokens_in, tokens_out, cost_usd) VALUES (?, ?, ?, ?, ?, ?)`,
		sessionID, provider, model, tokensIn, tokensOut, cost,
	)
	if err != nil {
		return fmt.Errorf("log usage: %w", err)
	}
	return nil
}

// GetSessionUsage returns the aggregated usage for a single session.
func (s *Store) GetSessionUsage(sessionID string) (UsageSummary, error) {
	if err := s.EnsureUsageTable(); err != nil {
		return UsageSummary{}, err
	}
	var u UsageSummary
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(tokens_in),0), COALESCE(SUM(tokens_out),0), COALESCE(SUM(cost_usd),0)
		 FROM usage_log WHERE session_id = ?`, sessionID,
	).Scan(&u.TokensIn, &u.TokensOut, &u.CostUSD)
	if err != nil {
		return u, fmt.Errorf("get session usage: %w", err)
	}
	return u, nil
}

// GetTotalUsage returns the aggregated usage across all sessions.
func (s *Store) GetTotalUsage() (UsageSummary, error) {
	if err := s.EnsureUsageTable(); err != nil {
		return UsageSummary{}, err
	}
	var u UsageSummary
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(tokens_in),0), COALESCE(SUM(tokens_out),0), COALESCE(SUM(cost_usd),0)
		 FROM usage_log`,
	).Scan(&u.TokensIn, &u.TokensOut, &u.CostUSD)
	if err != nil {
		return u, fmt.Errorf("get total usage: %w", err)
	}
	return u, nil
}

// FormatUsage returns a human-readable usage string.
func FormatUsage(u UsageSummary) string {
	total := u.TokensIn + u.TokensOut
	if total == 0 {
		return "사용량 없음"
	}
	if u.CostUSD > 0 {
		return fmt.Sprintf("토큰: %d (입력: %d, 출력: %d) | 비용: $%.4f",
			total, u.TokensIn, u.TokensOut, u.CostUSD)
	}
	return fmt.Sprintf("토큰: %d (입력: %d, 출력: %d)", total, u.TokensIn, u.TokensOut)
}
