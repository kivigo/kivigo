package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EventType represents the type of operation that triggered a hook.
type EventType string

const (
	// EventSet is triggered when Set operation is successful.
	EventSet EventType = "SET"
	// EventSetRaw is triggered when SetRaw operation is successful.
	EventSetRaw EventType = "SET_RAW"
	// EventDelete is triggered when Delete operation is successful.
	EventDelete EventType = "DELETE"
	// EventBatchSet is triggered when BatchSet operation is successful.
	EventBatchSet EventType = "BATCH_SET"
	// EventBatchDel is triggered when BatchDelete operation is successful.
	EventBatchDel EventType = "BATCH_DELETE"
)

// HookFunc is the function signature for hooks.
// It receives the context, event type, key, and value (if applicable).
// For delete operations, value will be nil.
type HookFunc func(ctx context.Context, evt EventType, key string, value []byte) error

// HookFilterFunc is a function that returns true if the hook should be executed for the given key.
type HookFilterFunc func(key string) bool

// HookOptions configures how a hook should be executed.
type HookOptions struct {
	// Events specifies which events this hook should respond to.
	// If nil or empty, the hook responds to all events.
	Events []EventType

	// Filter specifies a function to filter which keys this hook should respond to.
	// If nil, the hook responds to all keys.
	Filter HookFilterFunc

	// Async specifies whether the hook should be executed asynchronously.
	// If false, the hook is executed synchronously.
	Async bool

	// Timeout specifies the maximum time to wait for a synchronous hook to complete.
	// If zero, no timeout is applied.
	// This is ignored for async hooks.
	Timeout time.Duration
}

// hookRegistration represents a registered hook with its metadata.
type hookRegistration struct {
	id       string
	callback HookFunc
	options  HookOptions
	errCh    chan error
}

// HooksRegistry manages hook registration and execution.
type HooksRegistry struct {
	mu    sync.RWMutex
	hooks map[string]*hookRegistration
}

// NewHooksRegistry creates a new hooks registry.
func NewHooksRegistry() *HooksRegistry {
	return &HooksRegistry{
		hooks: make(map[string]*hookRegistration),
	}
}

// RegisterHook registers a new hook with the given callback and options.
// Returns a unique hook ID, an error channel for receiving hook errors,
// and an unregister function to remove the hook.
func (hr *HooksRegistry) RegisterHook(cb HookFunc, opts HookOptions) (string, <-chan error, func()) {
	id := generateHookID()
	errCh := make(chan error, 100) // Buffered channel for best-effort error delivery

	registration := &hookRegistration{
		id:       id,
		callback: cb,
		options:  opts,
		errCh:    errCh,
	}

	hr.mu.Lock()
	hr.hooks[id] = registration
	hr.mu.Unlock()

	unregister := func() {
		hr.UnregisterHook(id)
	}

	return id, errCh, unregister
}

// UnregisterHook removes a hook by its ID.
func (hr *HooksRegistry) UnregisterHook(id string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if registration, exists := hr.hooks[id]; exists {
		close(registration.errCh)
		delete(hr.hooks, id)
	}
}

// Run executes all registered hooks that match the given event and key.
// This method is called internally after successful operations.
func (hr *HooksRegistry) Run(ctx context.Context, evt EventType, key string, value []byte) {
	// Take a snapshot of hooks to avoid deadlocks
	hr.mu.RLock()
	snapshot := make([]*hookRegistration, 0, len(hr.hooks))
	for _, registration := range hr.hooks {
		if hr.shouldExecuteHook(registration, evt, key) {
			snapshot = append(snapshot, registration)
		}
	}
	hr.mu.RUnlock()

	// Execute hooks from snapshot
	for _, registration := range snapshot {
		if registration.options.Async {
			go hr.executeHookAsync(ctx, registration, evt, key, value)
		} else {
			hr.executeHookSync(ctx, registration, evt, key, value)
		}
	}
}

// shouldExecuteHook determines if a hook should be executed based on event type and key.
func (hr *HooksRegistry) shouldExecuteHook(registration *hookRegistration, evt EventType, key string) bool {
	// Check event filter
	if len(registration.options.Events) > 0 {
		found := false
		for _, allowedEvent := range registration.options.Events {
			if allowedEvent == evt {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check key filter
	if registration.options.Filter != nil && !registration.options.Filter(key) {
		return false
	}

	return true
}

// executeHookSync executes a hook synchronously with optional timeout.
func (hr *HooksRegistry) executeHookSync(ctx context.Context, registration *hookRegistration, evt EventType, key string, value []byte) {
	var execCtx context.Context
	var cancel context.CancelFunc

	if registration.options.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, registration.options.Timeout)
		defer cancel()
	} else {
		execCtx = ctx
	}

	err := registration.callback(execCtx, evt, key, value)
	if err != nil {
		// Best-effort error delivery
		select {
		case registration.errCh <- err:
		default:
			// Channel is full, drop the error
		}
	}
}

// executeHookAsync executes a hook asynchronously.
func (hr *HooksRegistry) executeHookAsync(ctx context.Context, registration *hookRegistration, evt EventType, key string, value []byte) {
	err := registration.callback(ctx, evt, key, value)
	if err != nil {
		// Best-effort error delivery
		select {
		case registration.errCh <- err:
		default:
			// Channel is full, drop the error
		}
	}
}

// generateHookID generates a unique ID for a hook.
func generateHookID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes) // Best-effort random ID generation
	return hex.EncodeToString(bytes)
}

// Common filter functions for convenience

// PrefixFilter returns a filter function that matches keys with the given prefix.
func PrefixFilter(prefix string) HookFilterFunc {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

// SuffixFilter returns a filter function that matches keys with the given suffix.
func SuffixFilter(suffix string) HookFilterFunc {
	return func(key string) bool {
		return strings.HasSuffix(key, suffix)
	}
}

// ListFilter returns a filter function that matches keys in the given list.
func ListFilter(keys []string) HookFilterFunc {
	keySet := make(map[string]bool, len(keys))
	for _, key := range keys {
		keySet[key] = true
	}
	return func(key string) bool {
		return keySet[key]
	}
}

// RegexFilter returns a filter function that matches keys against the given regex pattern.
func RegexFilter(pattern string) HookFilterFunc {
	regex := regexp.MustCompile(pattern)
	return func(key string) bool {
		return regex.MatchString(key)
	}
}
