package state

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockStateStore is an in-memory StateStore for testing.
type mockStateStore struct {
	mu       sync.Mutex
	cooldowns []CooldownEntry
	metrics   []TokenMetricsEntry
	stats     []RequestStatsEntry
	initErr   error
	saveErr   error
	loadErr   error
}

func (m *mockStateStore) Init(ctx context.Context) error { return m.initErr }
func (m *mockStateStore) Close() error                   { return nil }

func (m *mockStateStore) SaveCooldowns(_ context.Context, c []CooldownEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.cooldowns = make([]CooldownEntry, len(c))
	copy(m.cooldowns, c)
	return nil
}

func (m *mockStateStore) LoadCooldowns(_ context.Context) ([]CooldownEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	out := make([]CooldownEntry, len(m.cooldowns))
	copy(out, m.cooldowns)
	return out, nil
}

func (m *mockStateStore) SaveTokenMetrics(_ context.Context, met []TokenMetricsEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.metrics = make([]TokenMetricsEntry, len(met))
	copy(m.metrics, met)
	return nil
}

func (m *mockStateStore) LoadTokenMetrics(_ context.Context) ([]TokenMetricsEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	out := make([]TokenMetricsEntry, len(m.metrics))
	copy(out, m.metrics)
	return out, nil
}

func (m *mockStateStore) SaveRequestStats(_ context.Context, s []RequestStatsEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.stats = make([]RequestStatsEntry, len(s))
	copy(m.stats, s)
	return nil
}

func (m *mockStateStore) LoadRequestStats(_ context.Context) ([]RequestStatsEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	out := make([]RequestStatsEntry, len(m.stats))
	copy(out, m.stats)
	return out, nil
}

func TestSyncer_FlushAndLoad(t *testing.T) {
	store := &mockStateStore{}

	// Simulated in-memory state
	cooldowns := []CooldownEntry{
		{TokenKey: "tok-1", ExpiresAt: time.Now().Add(10 * time.Minute), Reason: "rate_limit"},
	}
	stats := []RequestStatsEntry{
		{StatDate: "2026-04-17", TotalRequests: 100, SuccessCount: 90, FailureCount: 10, TotalTokens: 5000},
	}

	var imported []CooldownEntry
	var importedStats []RequestStatsEntry

	syncer := NewSyncer(store, SyncerConfig{FlushInterval: 50 * time.Millisecond})
	syncer.ExportCooldowns = func() []CooldownEntry { return cooldowns }
	syncer.ImportCooldowns = func(e []CooldownEntry) { imported = e }
	syncer.ExportStats = func() []RequestStatsEntry { return stats }
	syncer.ImportStats = func(e []RequestStatsEntry) { importedStats = e }

	// Start and wait for at least one flush
	syncer.Start()
	time.Sleep(120 * time.Millisecond)
	syncer.Stop()

	// Verify data was flushed to store
	store.mu.Lock()
	if len(store.cooldowns) != 1 {
		t.Errorf("expected 1 cooldown in store, got %d", len(store.cooldowns))
	}
	if len(store.stats) != 1 {
		t.Errorf("expected 1 stat in store, got %d", len(store.stats))
	}
	store.mu.Unlock()

	// Now simulate restart: load state back
	syncer2 := NewSyncer(store, SyncerConfig{FlushInterval: time.Hour})
	syncer2.ImportCooldowns = func(e []CooldownEntry) { imported = e }
	syncer2.ImportStats = func(e []RequestStatsEntry) { importedStats = e }

	syncer2.LoadState(context.Background())

	if len(imported) != 1 {
		t.Errorf("expected 1 imported cooldown, got %d", len(imported))
	}
	if imported[0].TokenKey != "tok-1" {
		t.Errorf("expected token key tok-1, got %s", imported[0].TokenKey)
	}
	if len(importedStats) != 1 {
		t.Errorf("expected 1 imported stat, got %d", len(importedStats))
	}
}

func TestSyncer_NilExporters(t *testing.T) {
	store := &mockStateStore{}
	syncer := NewSyncer(store, SyncerConfig{FlushInterval: 50 * time.Millisecond})

	// No exporters set — should not panic
	syncer.Start()
	time.Sleep(80 * time.Millisecond)
	syncer.Stop()

	syncer2 := NewSyncer(store, SyncerConfig{})
	syncer2.LoadState(context.Background()) // should not panic
}

func TestSyncer_DefaultFlushInterval(t *testing.T) {
	syncer := NewSyncer(&mockStateStore{}, SyncerConfig{})
	if syncer.cfg.FlushInterval != 30*time.Second {
		t.Errorf("expected default flush interval 30s, got %s", syncer.cfg.FlushInterval)
	}
}

func TestSyncer_GracefulShutdown(t *testing.T) {
	store := &mockStateStore{}
	flushCount := 0

	syncer := NewSyncer(store, SyncerConfig{FlushInterval: time.Hour}) // long interval
	syncer.ExportCooldowns = func() []CooldownEntry {
		flushCount++
		return []CooldownEntry{{TokenKey: "tok", ExpiresAt: time.Now().Add(time.Hour), Reason: "test"}}
	}

	syncer.Start()
	// Stop immediately — should trigger exactly one final flush
	syncer.Stop()

	if flushCount != 1 {
		t.Errorf("expected 1 final flush on stop, got %d", flushCount)
	}
	store.mu.Lock()
	if len(store.cooldowns) != 1 {
		t.Errorf("expected 1 cooldown after final flush, got %d", len(store.cooldowns))
	}
	store.mu.Unlock()
}
