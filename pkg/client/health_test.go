package client

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockKVWithoutHealth struct{}

type mockKVWithHealth struct {
	mockKVWithoutHealth
	healthErr error
}

func (m *mockKVWithHealth) Health(_ context.Context) error {
	return m.healthErr
}

// Satisfait l'interface models.KV
func (m *mockKVWithoutHealth) GetRaw(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (m *mockKVWithoutHealth) SetRaw(_ context.Context, _ string, _ []byte) error { return nil }
func (m *mockKVWithoutHealth) Delete(_ context.Context, _ string) error           { return nil }
func (m *mockKVWithoutHealth) List(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockKVWithoutHealth) Close() error { return nil }

// Mock pour Client
func newMockClient(healthErr error) Client {
	return Client{
		KV: &mockKVWithHealth{healthErr: healthErr},
	}
}

func TestHealth_NoKVWithHealth(t *testing.T) {
	c := Client{KV: &mockKVWithoutHealth{}}

	err := c.Health(t.Context(), nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHealth_KVWithHealthHealthy(t *testing.T) {
	c := newMockClient(nil)

	err := c.Health(t.Context(), nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHealth_KVWithHealthUnhealthy(t *testing.T) {
	c := newMockClient(errors.New("backend down"))

	err := c.Health(t.Context(), nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestHealth_WithAdditionalChecks_AllHealthy(t *testing.T) {
	c := newMockClient(nil)
	check := func(_ context.Context, _ Client) error { return nil }

	err := c.Health(t.Context(), []HealthFunc{check, check})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHealth_WithAdditionalChecks_OneFails(t *testing.T) {
	c := newMockClient(nil)
	okCheck := func(_ context.Context, _ Client) error { return nil }
	failCheck := func(_ context.Context, _ Client) error { return errors.New("fail") }

	err := c.Health(t.Context(), []HealthFunc{okCheck, failCheck})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestHealthCheck_Healthy(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	c := newMockClient(nil)
	ho := HealthOptions{Interval: 10 * time.Millisecond}
	ch := c.HealthCheck(ctx, ho)

	select {
	case err := <-ch:
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		return
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	c := newMockClient(errors.New("backend down"))
	ho := HealthOptions{Interval: 10 * time.Millisecond}
	ch := c.HealthCheck(ctx, ho)

	select {
	case err := <-ch:
		if err == nil {
			t.Error("expected error, got nil")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for health check")
	}
}

func TestHealthCheck_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	c := newMockClient(nil)
	ho := HealthOptions{Interval: 50 * time.Millisecond}
	ch := c.HealthCheck(ctx, ho)

	cancel()

	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed after context cancel")
	}
}

// func TestHealthCheck_DefaultInterval(t *testing.T) {
// 	ctx, cancel := context.WithCancel(t.Context())
// 	defer cancel()
// 	c := newMockClient(nil)
// 	ho := HealthOptions{Interval: 0}
// 	ch := c.HealthCheck(ctx, ho)

// 	select {
// 	case <-ch:
// 		// ok, should emit at least one value
// 	case <-time.After(100 * time.Millisecond):
// 	}
// }
