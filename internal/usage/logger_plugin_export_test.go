package usage

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/state"
)

func TestRequestStatistics_ExportImportDailyStats(t *testing.T) {
	s := NewRequestStatistics()

	s.mu.Lock()
	s.totalRequests = 100
	s.successCount = 90
	s.failureCount = 10
	s.totalTokens = 5000
	s.requestsByDay["2026-04-17"] = 80
	s.requestsByDay["2026-04-16"] = 20
	s.tokensByDay["2026-04-17"] = 4000
	s.tokensByDay["2026-04-16"] = 1000
	s.mu.Unlock()

	entries := s.ExportDailyStats()
	// 2 day entries + 1 _global entry
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (2 days + 1 global), got %d", len(entries))
	}

	s2 := NewRequestStatistics()
	s2.ImportDailyStats(entries)

	s2.mu.RLock()
	defer s2.mu.RUnlock()

	if s2.requestsByDay["2026-04-17"] != 80 {
		t.Errorf("expected 80 requests for 2026-04-17, got %d", s2.requestsByDay["2026-04-17"])
	}
	if s2.tokensByDay["2026-04-17"] != 4000 {
		t.Errorf("expected 4000 tokens for 2026-04-17, got %d", s2.tokensByDay["2026-04-17"])
	}
	// Global counters should be exact (from _global sentinel), not approximated
	if s2.totalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", s2.totalRequests)
	}
	if s2.successCount != 90 {
		t.Errorf("expected 90 success count, got %d", s2.successCount)
	}
	if s2.failureCount != 10 {
		t.Errorf("expected 10 failure count, got %d", s2.failureCount)
	}
	if s2.totalTokens != 5000 {
		t.Errorf("expected 5000 total tokens, got %d", s2.totalTokens)
	}
}

func TestRequestStatistics_ExportDailyStats_Empty(t *testing.T) {
	s := NewRequestStatistics()
	entries := s.ExportDailyStats()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries from empty stats, got %d", len(entries))
	}
}

func TestRequestStatistics_ImportNil(t *testing.T) {
	s := NewRequestStatistics()
	s.mu.Lock()
	s.totalRequests = 5
	s.mu.Unlock()

	s.ImportDailyStats(nil)

	s.mu.RLock()
	defer s.mu.RUnlock()
	// Import with nil has no _global entry, fallback sums to 0
	if s.totalRequests != 0 {
		t.Errorf("expected 0 total requests after nil import, got %d", s.totalRequests)
	}
}

func TestRequestStatistics_ExportReturnsStateType(t *testing.T) {
	s := NewRequestStatistics()
	s.mu.Lock()
	s.requestsByDay["2026-04-17"] = 10
	s.totalRequests = 10
	s.successCount = 10
	s.tokensByDay["2026-04-17"] = 100
	s.mu.Unlock()

	entries := s.ExportDailyStats()
	// Verify type compatibility
	var _ []state.RequestStatsEntry = entries
	// 1 day + 1 _global
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestRequestStatistics_ImportFallbackNoGlobal(t *testing.T) {
	// Test import without _global sentinel (backwards compatibility)
	s := NewRequestStatistics()
	entries := []state.RequestStatsEntry{
		{StatDate: "2026-04-17", TotalRequests: 50, TotalTokens: 2000},
		{StatDate: "2026-04-16", TotalRequests: 30, TotalTokens: 1000},
	}
	s.ImportDailyStats(entries)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.totalRequests != 80 {
		t.Errorf("expected 80 total requests from fallback sum, got %d", s.totalRequests)
	}
	if s.totalTokens != 3000 {
		t.Errorf("expected 3000 total tokens from fallback sum, got %d", s.totalTokens)
	}
}

func TestRequestStatistics_GlobalSentinelNotInDayMap(t *testing.T) {
	s := NewRequestStatistics()
	s.mu.Lock()
	s.totalRequests = 10
	s.successCount = 10
	s.requestsByDay["2026-04-17"] = 10
	s.tokensByDay["2026-04-17"] = 100
	s.mu.Unlock()

	entries := s.ExportDailyStats()
	s2 := NewRequestStatistics()
	s2.ImportDailyStats(entries)

	s2.mu.RLock()
	defer s2.mu.RUnlock()
	// _global sentinel should NOT appear in requestsByDay
	if _, exists := s2.requestsByDay["_global"]; exists {
		t.Error("_global sentinel should not be stored in requestsByDay")
	}
}
