package grpcserver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"remote_judge/internal/domain"
	"remote_judge/internal/judger"
	"remote_judge/internal/transport/grpcjson"
	"remote_judge/pkg/pb"
)

// Server 封装 Judger gRPC 服务。
type Server struct {
	judgeSvc *judger.Service
}

// JudgerService 定义 gRPC 服务需要满足的接口。
type JudgerService interface {
	Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error)
	Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error)
}

// New 创建一个 gRPC Judger 服务处理器。
func New(judgeSvc *judger.Service) *Server {
	return &Server{judgeSvc: judgeSvc}
}

// Register 将服务注册到 gRPC server。
func Register(s *grpc.Server, handler *Server) {
	encoding.RegisterCodec(grpcjson.Codec{})
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "judger.Judger",
		HandlerType: (*JudgerService)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Judge",
				Handler:    handler.judgeHandler,
			},
			{
				MethodName: "Health",
				Handler:    handler.healthHandler,
			},
		},
	}, handler)
}

// Judge 处理一次 gRPC 判题请求。
func (s *Server) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	testCases := make([]domain.TestCase, 0, len(req.TestCases))
	for _, tc := range req.TestCases {
		testCases = append(testCases, domain.TestCase{
			ProblemID: req.ProblemID,
			CaseNo:    int(tc.CaseNo),
			Input:     tc.Input,
			Expected:  tc.Expected,
		})
	}
	result, err := s.judgeSvc.Judge(ctx, domain.JudgeRequest{
		SubmissionID:  req.SubmissionID,
		ProblemID:     req.ProblemID,
		Language:      req.Language,
		Code:          req.Code,
		TimeLimitMs:   int(req.TimeLimitMs),
		MemoryLimitMB: int(req.MemoryLimitMB),
		OutputLimitKB: int(req.OutputLimitKB),
		TestCases:     testCases,
	})
	if err != nil {
		return nil, err
	}
	resp := &pb.JudgeResponse{
		SubmissionID: result.SubmissionID,
		Status:       string(result.Status),
		RuntimeMs:    int32(result.RuntimeMs),
		MemoryKB:     int32(result.MemoryKB),
		CompileOut:   result.CompileOut,
		ErrorMessage: result.ErrorMessage,
	}
	for _, item := range result.CaseResults {
		resp.CaseResults = append(resp.CaseResults, pb.CaseResult{
			CaseNo:        int32(item.CaseNo),
			Status:        string(item.Status),
			RuntimeMs:     int32(item.RuntimeMs),
			MemoryKB:      int32(item.MemoryKB),
			StdoutPreview: item.StdoutPreview,
			StderrPreview: item.StderrPreview,
		})
	}
	return resp, nil
}

// Health 处理 gRPC 健康检查请求。
func (s *Server) Health(ctx context.Context, _ *pb.HealthRequest) (*pb.HealthResponse, error) {
	err := s.judgeSvc.Health(ctx)
	langs := make([]string, 0, len(domain.SupportedLanguages))
	for key := range domain.SupportedLanguages {
		langs = append(langs, key)
	}
	return &pb.HealthResponse{
		Status:             "SERVING",
		DockerReady:        err == nil,
		SupportedLanguages: langs,
	}, nil
}

// judgeHandler 适配 gRPC method descriptor。
func (s *Server) judgeHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := new(pb.JudgeRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(*Server).Judge(ctx, req)
}

// healthHandler 适配 gRPC method descriptor。
func (s *Server) healthHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := new(pb.HealthRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(*Server).Health(ctx, req)
}
