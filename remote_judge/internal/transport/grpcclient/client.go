package grpcclient

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"

	"remote_judge/internal/domain"
	"remote_judge/internal/transport/grpcjson"
	"remote_judge/pkg/pb"
)

// Client 提供远程 Judger gRPC 调用能力。
type Client struct {
	addr string
	conn *grpc.ClientConn
}

// New 创建一个远程 Judger 客户端。
func New(addr string) (*Client, error) {
	encoding.RegisterCodec(grpcjson.Codec{})
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype("json")),
	)
	if err != nil {
		return nil, err
	}
	return &Client{addr: addr, conn: conn}, nil
}

// Close 关闭底层连接。
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Judge 调用远程 Judger 服务。
func (c *Client) Judge(ctx context.Context, req domain.JudgeRequest) (domain.JudgeResult, error) {
	if c.conn == nil {
		return domain.JudgeResult{}, errors.New("grpc client not initialized")
	}
	testCases := make([]pb.TestCase, 0, len(req.TestCases))
	for _, tc := range req.TestCases {
		testCases = append(testCases, pb.TestCase{
			CaseNo:   int32(tc.CaseNo),
			Input:    tc.Input,
			Expected: tc.Expected,
		})
	}
	resp := new(pb.JudgeResponse)
	err := c.conn.Invoke(ctx, "/judger.Judger/Judge", &pb.JudgeRequest{
		SubmissionID:  req.SubmissionID,
		ProblemID:     req.ProblemID,
		Language:      req.Language,
		Code:          req.Code,
		TimeLimitMs:   int32(req.TimeLimitMs),
		MemoryLimitMB: int32(req.MemoryLimitMB),
		OutputLimitKB: int32(req.OutputLimitKB),
		TestCases:     testCases,
	}, resp)
	if err != nil {
		return domain.JudgeResult{}, err
	}
	result := domain.JudgeResult{
		SubmissionID: resp.SubmissionID,
		Status:       domain.SubmissionStatus(resp.Status),
		RuntimeMs:    int(resp.RuntimeMs),
		MemoryKB:     int(resp.MemoryKB),
		CompileOut:   resp.CompileOut,
		ErrorMessage: resp.ErrorMessage,
	}
	for _, item := range resp.CaseResults {
		result.CaseResults = append(result.CaseResults, domain.SubmissionCaseResult{
			SubmissionID:  resp.SubmissionID,
			CaseNo:        int(item.CaseNo),
			Status:        domain.SubmissionStatus(item.Status),
			RuntimeMs:     int(item.RuntimeMs),
			MemoryKB:      int(item.MemoryKB),
			StdoutPreview: item.StdoutPreview,
			StderrPreview: item.StderrPreview,
		})
	}
	return result, nil
}

// Health 调用远程 Judger 健康检查。
func (c *Client) Health(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("grpc client not initialized")
	}
	resp := new(pb.HealthResponse)
	return c.conn.Invoke(ctx, "/judger.Judger/Health", &pb.HealthRequest{}, resp)
}
