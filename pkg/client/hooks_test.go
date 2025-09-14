package client

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kivigo/encoders/json"

	"github.com/kivigo/kivigo/pkg/mock"
)

func TestHookRegistry_RegisterAndUnregister(t *testing.T) {
	registry := NewHooksRegistry()

	hookCalled := false
	hookFunc := func(ctx context.Context, evt EventType, key string, value []byte) error {
		hookCalled = true
		return nil
	}

	// Register hook
	id, errCh, unregister := registry.RegisterHook(hookFunc, HookOptions{})
	if id == "" {
		t.Error("Expected non-empty hook ID")
	}

	// Trigger hook
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))

	// Check if hook was called
	if !hookCalled {
		t.Error("Hook should have been called")
	}

	// Reset and unregister
	hookCalled = false
	unregister()

	// Trigger again - should not be called
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))
	if hookCalled {
		t.Error("Hook should not have been called after unregister")
	}

	// Check error channel is closed
	select {
	case _, ok := <-errCh:
		if ok {
			t.Error("Error channel should be closed after unregister")
		}
	default:
		t.Error("Error channel should be closed after unregister")
	}
}

func TestHookRegistry_EventFilter(t *testing.T) {
	registry := NewHooksRegistry()

	setCalled := false
	deleteCalled := false

	setHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		setCalled = true
		return nil
	}

	deleteHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		deleteCalled = true
		return nil
	}

	// Register hooks with event filters
	_, _, unregisterSet := registry.RegisterHook(setHook, HookOptions{
		Events: []EventType{EventSet},
	})
	defer unregisterSet()

	_, _, unregisterDelete := registry.RegisterHook(deleteHook, HookOptions{
		Events: []EventType{EventDelete},
	})
	defer unregisterDelete()

	// Trigger SET event
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))
	if !setCalled {
		t.Error("Set hook should have been called")
	}
	if deleteCalled {
		t.Error("Delete hook should not have been called")
	}

	// Reset and trigger DELETE event
	setCalled = false
	deleteCalled = false
	registry.Run(context.Background(), EventDelete, "test-key", nil)
	if setCalled {
		t.Error("Set hook should not have been called")
	}
	if !deleteCalled {
		t.Error("Delete hook should have been called")
	}
}

func TestHookRegistry_KeyFilters(t *testing.T) {
	registry := NewHooksRegistry()

	prefixCalled := false
	suffixCalled := false
	listCalled := false
	regexCalled := false

	hookFunc := func(called *bool) HookFunc {
		return func(ctx context.Context, evt EventType, key string, value []byte) error {
			*called = true
			return nil
		}
	}

	// Register hooks with different filters
	_, _, unregisterPrefix := registry.RegisterHook(hookFunc(&prefixCalled), HookOptions{
		Filter: PrefixFilter("user:"),
	})
	defer unregisterPrefix()

	_, _, unregisterSuffix := registry.RegisterHook(hookFunc(&suffixCalled), HookOptions{
		Filter: SuffixFilter(":config"),
	})
	defer unregisterSuffix()

	_, _, unregisterList := registry.RegisterHook(hookFunc(&listCalled), HookOptions{
		Filter: ListFilter([]string{"key1", "key2", "key3"}),
	})
	defer unregisterList()

	_, _, unregisterRegex := registry.RegisterHook(hookFunc(&regexCalled), HookOptions{
		Filter: RegexFilter(`^session:[a-f0-9]+$`),
	})
	defer unregisterRegex()

	// Test prefix filter
	registry.Run(context.Background(), EventSet, "user:123", []byte("data"))
	if !prefixCalled {
		t.Error("Prefix hook should have been called")
	}

	// Test suffix filter
	registry.Run(context.Background(), EventSet, "app:config", []byte("data"))
	if !suffixCalled {
		t.Error("Suffix hook should have been called")
	}

	// Test list filter
	registry.Run(context.Background(), EventSet, "key2", []byte("data"))
	if !listCalled {
		t.Error("List hook should have been called")
	}

	// Test regex filter
	registry.Run(context.Background(), EventSet, "session:abc123", []byte("data"))
	if !regexCalled {
		t.Error("Regex hook should have been called")
	}

	// Reset and test non-matching keys
	prefixCalled = false
	suffixCalled = false
	listCalled = false
	regexCalled = false

	registry.Run(context.Background(), EventSet, "other:key", []byte("data"))
	if prefixCalled || suffixCalled || listCalled || regexCalled {
		t.Error("No hooks should have been called for non-matching key")
	}
}

func TestHookRegistry_SyncAndAsync(t *testing.T) {
	registry := NewHooksRegistry()

	syncCalled := false
	asyncCalled := false

	var syncMu, asyncMu sync.Mutex

	syncHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		syncMu.Lock()
		defer syncMu.Unlock()
		syncCalled = true
		return nil
	}

	asyncHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		asyncMu.Lock()
		defer asyncMu.Unlock()
		asyncCalled = true
		return nil
	}

	// Register sync and async hooks
	_, _, unregisterSync := registry.RegisterHook(syncHook, HookOptions{Async: false})
	defer unregisterSync()

	_, _, unregisterAsync := registry.RegisterHook(asyncHook, HookOptions{Async: true})
	defer unregisterAsync()

	// Trigger hooks
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))

	// Sync hook should be called immediately
	syncMu.Lock()
	syncCallResult := syncCalled
	syncMu.Unlock()

	if !syncCallResult {
		t.Error("Sync hook should have been called immediately")
	}

	// Async hook might need a moment
	time.Sleep(10 * time.Millisecond)

	asyncMu.Lock()
	asyncCallResult := asyncCalled
	asyncMu.Unlock()

	if !asyncCallResult {
		t.Error("Async hook should have been called")
	}
}

func TestHookRegistry_ErrorHandling(t *testing.T) {
	registry := NewHooksRegistry()

	expectedErr := errors.New("hook error")

	errorHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		return expectedErr
	}

	// Register hook that returns error
	_, errCh, unregister := registry.RegisterHook(errorHook, HookOptions{})
	defer unregister()

	// Trigger hook
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))

	// Check error is received
	select {
	case err := <-errCh:
		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive error from hook")
	}
}

func TestHookRegistry_TimeoutHandling(t *testing.T) {
	registry := NewHooksRegistry()

	slowHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Register hook with timeout
	_, errCh, unregister := registry.RegisterHook(slowHook, HookOptions{
		Timeout: 50 * time.Millisecond,
	})
	defer unregister()

	// Trigger hook
	registry.Run(context.Background(), EventSet, "test-key", []byte("test-value"))

	// Check timeout error is received
	select {
	case err := <-errCh:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected DeadlineExceeded error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive timeout error from hook")
	}
}

func TestClient_HookIntegration(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{}}
	c, err := New(mockKV, Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	var events []string
	var mu sync.Mutex

	hookFunc := func(ctx context.Context, evt EventType, key string, value []byte) error {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, string(evt)+" "+key)
		return nil
	}

	// Register hook
	_, _, unregister := c.RegisterHook(hookFunc, HookOptions{})
	defer unregister()

	ctx := context.Background()

	// Test Set operation
	err = c.Set(ctx, "test-key", "test-value")
	if err != nil {
		t.Fatal(err)
	}

	// Test Delete operation
	err = c.Delete(ctx, "test-key")
	if err != nil {
		t.Fatal(err)
	}

	// Test BatchSet operation
	err = c.BatchSet(ctx, map[string]any{"key1": "value1", "key2": "value2"})
	if err != nil {
		t.Fatal(err)
	}

	// Test BatchDelete operation
	err = c.BatchDelete(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Fatal(err)
	}

	// Check events
	mu.Lock()
	defer mu.Unlock()

	expectedEvents := []string{
		"SET test-key",
		"DELETE test-key",
		"BATCH_SET key1",
		"BATCH_SET key2",
		"BATCH_DELETE key1",
		"BATCH_DELETE key2",
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("Expected %d events, got %d: %v", len(expectedEvents), len(events), events)
	}

	// Note: BatchSet events might come in different order due to map iteration
	// So we check if all expected events are present
	eventSet := make(map[string]bool)
	for _, event := range events {
		eventSet[event] = true
	}

	for _, expected := range expectedEvents {
		if !eventSet[expected] {
			t.Errorf("Expected event %s not found in %v", expected, events)
		}
	}
}

func TestClient_HookErrorsDoNotFailOperations(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{}}
	c, err := New(mockKV, Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	// Register a hook that always fails
	errorHook := func(ctx context.Context, evt EventType, key string, value []byte) error {
		return errors.New("hook error")
	}

	_, errCh, unregister := c.RegisterHook(errorHook, HookOptions{})
	defer unregister()

	ctx := context.Background()

	// Operations should succeed despite hook errors
	err = c.Set(ctx, "test-key", "test-value")
	if err != nil {
		t.Errorf("Set operation should succeed even with hook errors, got: %v", err)
	}

	err = c.Delete(ctx, "test-key")
	if err != nil {
		t.Errorf("Delete operation should succeed even with hook errors, got: %v", err)
	}

	// Check that errors are reported via error channel
	timeout := time.After(100 * time.Millisecond)
	errorCount := 0

	for errorCount < 2 {
		select {
		case <-errCh:
			errorCount++
		case <-timeout:
			t.Errorf("Expected to receive 2 errors from hooks, got %d", errorCount)
			return
		}
	}
}
