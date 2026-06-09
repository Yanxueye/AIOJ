package sandbox

import (
	"context"
	"errors"
	"sync"
	"time"

	"remote_judge/internal/logger"
)

var errSandboxUnavailable = errors.New("judge sandbox is temporarily unavailable")

type breakerState string

const (
	breakerClosed   breakerState = "closed"
	breakerOpen     breakerState = "open"
	breakerHalfOpen breakerState = "half_open"
)

// CircuitBreakerSandbox 通过进程内轻量熔断器包装沙箱。
type CircuitBreakerSandbox struct {
	next             Sandbox
	failureThreshold int
	halfOpenAfter    time.Duration

	mu             sync.Mutex
	state          breakerState
	consecutiveErr int
	openedAt       time.Time
	probeRunning   bool
}

// NewCircuitBreaker 创建带熔断器保护的沙箱。
func NewCircuitBreaker(next Sandbox) *CircuitBreakerSandbox {
	return &CircuitBreakerSandbox{
		next:             next,
		failureThreshold: 3,
		halfOpenAfter:    30 * time.Second,
		state:            breakerClosed,
	}
}

// Compile 在熔断器保护下执行编译。
func (c *CircuitBreakerSandbox) Compile(ctx context.Context, req ExecRequest) (ExecResult, error) {
	if err := c.before("compile"); err != nil {
		return ExecResult{}, err
	}
	res, err := c.next.Compile(ctx, req)
	c.after("compile", err)
	return res, err
}

// Run 在熔断器保护下执行运行。
func (c *CircuitBreakerSandbox) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	if err := c.before("run"); err != nil {
		return ExecResult{}, err
	}
	res, err := c.next.Run(ctx, req)
	c.after("run", err)
	return res, err
}

// Health 将健康检查委托给被包装的沙箱。
func (c *CircuitBreakerSandbox) Health(ctx context.Context) error {
	return c.next.Health(ctx)
}

func (c *CircuitBreakerSandbox) before(op string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	switch c.state {
	case breakerOpen:
		if now.Sub(c.openedAt) >= c.halfOpenAfter {
			c.state = breakerHalfOpen
			c.probeRunning = false
			logger.Info("sandbox.breaker", "", "breaker transitioned to half-open", map[string]any{"operation": op})
		} else {
			return errSandboxUnavailable
		}
	}

	if c.state == breakerHalfOpen {
		if c.probeRunning {
			return errSandboxUnavailable
		}
		c.probeRunning = true
	}
	return nil
}

func (c *CircuitBreakerSandbox) after(op string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err == nil {
		if c.state == breakerHalfOpen {
			c.state = breakerClosed
			c.consecutiveErr = 0
			c.probeRunning = false
			logger.Info("sandbox.breaker", "", "breaker closed after successful probe", map[string]any{"operation": op})
			return
		}
		c.consecutiveErr = 0
		return
	}

	switch c.state {
	case breakerHalfOpen:
		c.state = breakerOpen
		c.openedAt = time.Now()
		c.consecutiveErr = c.failureThreshold
		c.probeRunning = false
		logger.Error("sandbox.breaker", "", "breaker reopened after failed probe", map[string]any{"operation": op, "error": err.Error()})
	case breakerClosed:
		c.consecutiveErr++
		if c.consecutiveErr >= c.failureThreshold {
			c.state = breakerOpen
			c.openedAt = time.Now()
			logger.Error("sandbox.breaker", "", "breaker opened after consecutive failures", map[string]any{"operation": op, "error": err.Error(), "failures": c.consecutiveErr})
		}
	}
}
