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
	if len(entries) != 2 {
		t.Fatalf("expected 2 daily entries, got %d", len(entries))
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
	if s2.totalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", s2.totalRequests)
	}
	if s2.successCount != 90 {
		t.Errorf("expected 90 success count, got %d", s2.successCount)
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
	// Import with nil resets counters to 0 since there are no entries
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
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].StatDate != "2026-04-17" {
		t.Errorf("expected date 2026-04-17, got %s", entries[0].StatDate)
	}
}
