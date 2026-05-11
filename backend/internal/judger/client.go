package judger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a thin wrapper around *grpc.ClientConn that speaks our JSON
// codec. A single Client is safe for concurrent use by many goroutines.
type Client struct {
	addr    string
	timeout time.Duration

	mu   sync.Mutex
	conn *grpc.ClientConn
}

// NewClient lazily dials when the first Judge call is issued; this keeps
// the backend bootable even when the judger container is slow to start.
func NewClient(addr string, timeoutSeconds int) *Client {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}
	return &Client{addr: addr, timeout: time.Duration(timeoutSeconds) * time.Second}
}

func (c *Client) ensureConn() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return nil
	}
	conn, err := grpc.NewClient(c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codecName)),
	)
	if err != nil {
		return fmt.Errorf("grpc dial %s: %w", c.addr, err)
	}
	c.conn = conn
	return nil
}

// Close releases the underlying grpc.ClientConn.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Judge issues a blocking RPC to the remote judger container.
func (c *Client) Judge(ctx context.Context, req *JudgeRequest) (*JudgeResponse, error) {
	if err := c.ensureConn(); err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp := new(JudgeResponse)
	if err := c.conn.Invoke(callCtx, MethodJudge, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
