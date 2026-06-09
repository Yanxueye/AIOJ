package sandbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSandbox struct {
	err     error
	calls   int
	runOnly bool
}

func (f *fakeSandbox) Compile(context.Context, ExecRequest) (ExecResult, error) {
	f.calls++
	if f.runOnly {
		return ExecResult{}, nil
	}
	return ExecResult{}, f.err
}

func (f *fakeSandbox) Run(context.Context, ExecRequest) (ExecResult, error) {
	f.calls++
	return ExecResult{}, f.err
}

func (f *fakeSandbox) Health(context.Context) error { return nil }

// TestCircuitBreakerOpensAfterFailures verifies breaker opening after threshold failures.
func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	t.Logf(">>> Circuit Breaker: 3 failures -> Open state")
	base := &fakeSandbox{err: errors.New("boom")}
	cb := NewCircuitBreaker(base)
	for i := 0; i < 3; i++ {
		_, _ = cb.Run(context.Background(), ExecRequest{Command: []string{"x"}})
	}
	_, err := cb.Run(context.Background(), ExecRequest{Command: []string{"x"}})
	if !errors.Is(err, errSandboxUnavailable) {
		t.Fatalf("expected sandbox unavailable, got %v", err)
	}
	if base.calls != 3 {
		t.Fatalf("expected 3 delegated calls, got %d", base.calls)
	}
	t.Logf("    state=Open | delegated=%d | 4th request rejected", base.calls)
}

// TestCircuitBreakerHalfOpenRecovery verifies half-open probe success closes the breaker.
func TestCircuitBreakerHalfOpenRecovery(t *testing.T) {
	t.Logf(">>> Circuit Breaker: Open -> HalfOpen (probe) -> Closed (recovery)")
	base := &fakeSandbox{err: errors.New("boom")}
	cb := NewCircuitBreaker(base)
	cb.halfOpenAfter = 10 * time.Millisecond
	for i := 0; i < 3; i++ {
		_, _ = cb.Run(context.Background(), ExecRequest{Command: []string{"x"}})
	}
	time.Sleep(15 * time.Millisecond)
	base.err = nil
	if _, err := cb.Run(context.Background(), ExecRequest{Command: []string{"x"}}); err != nil {
		t.Fatalf("expected probe success, got %v", err)
	}
	if _, err := cb.Run(context.Background(), ExecRequest{Command: []string{"x"}}); err != nil {
		t.Fatalf("expected closed breaker, got %v", err)
	}
	t.Logf("    state=Closed | recovery probe succeeded -> full recovery")
}

// TestCircuitBreakerHalfOpenSingleProbe verifies only one probe runs in half-open state.
func TestCircuitBreakerHalfOpenSingleProbe(t *testing.T) {
	t.Logf(">>> Circuit Breaker: HalfOpen -> single probe (reject others)")
	base := &fakeSandbox{err: errors.New("boom")}
	cb := NewCircuitBreaker(base)
	cb.halfOpenAfter = 10 * time.Millisecond
	for i := 0; i < 3; i++ {
		_, _ = cb.Run(context.Background(), ExecRequest{Command: []string{"x"}})
	}
	time.Sleep(15 * time.Millisecond)

	cb.mu.Lock()
	cb.state = breakerHalfOpen
	cb.probeRunning = true
	cb.mu.Unlock()

	_, err := cb.Run(context.Background(), ExecRequest{Command: []string{"x"}})
	if !errors.Is(err, errSandboxUnavailable) {
		t.Fatalf("expected sandbox unavailable during active probe, got %v", err)
	}
	t.Logf("    state=HalfOpen(probeRunning) | concurrent request rejected")
}
