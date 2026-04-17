package state

import (
	"context"
	"time"
)

// StateStore defines the persistence interface for runtime state.
type StateStore interface {
	// Init creates required tables/schema.
	Init(ctx context.Context) error

	// Cooldowns
	SaveCooldowns(ctx context.Context, cooldowns []CooldownEntry) error
	LoadCooldowns(ctx context.Context) ([]CooldownEntry, error)

	// Token Metrics
	SaveTokenMetrics(ctx context.Context, metrics []TokenMetricsEntry) error
	LoadTokenMetrics(ctx context.Context) ([]TokenMetricsEntry, error)

	// Request Stats
	SaveRequestStats(ctx context.Context, stats []RequestStatsEntry) error
	LoadRequestStats(ctx context.Context) ([]RequestStatsEntry, error)

	// Close releases resources.
	Close() error
}

// CooldownEntry represents a persisted cooldown record.
type CooldownEntry struct {
	TokenKey  string
	ExpiresAt time.Time
	Reason    string
}

// TokenMetricsEntry represents persisted token performance metrics.
type TokenMetricsEntry struct {
	TokenKey       string
	SuccessRate    float64
	AvgLatency     float64
	QuotaRemaining float64
	LastUsed       time.Time
	FailCount      int
	TotalRequests  int
	SuccessCount   int
	TotalLatency   float64
}

// RequestStatsEntry represents persisted daily request statistics.
type RequestStatsEntry struct {
	StatDate      string // "2006-01-02" format
	TotalRequests int64
	SuccessCount  int64
	FailureCount  int64
	TotalTokens   int64
}
