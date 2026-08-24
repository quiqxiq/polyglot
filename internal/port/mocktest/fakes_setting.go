package mocktest

import (
	"context"
	"sync"
)

// ─── SettingReader ──────────────────────────────────────────────────────

type FakeSettingReader struct {
	mu   sync.Mutex
	vals map[string]string
}

func NewFakeSettingReader(vals map[string]string) *FakeSettingReader {
	out := map[string]string{}
	for k, v := range vals {
		out[k] = v
	}
	return &FakeSettingReader{vals: out}
}

func (f *FakeSettingReader) GetValue(_ context.Context, key, fallback string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if v, ok := f.vals[key]; ok && v != "" {
		return v
	}
	return fallback
}
