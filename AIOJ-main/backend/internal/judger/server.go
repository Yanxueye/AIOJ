package judger

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strings"

	"google.golang.org/grpc"
)

// Handler is the server-side implementation of the Judge RPC.
type Handler interface {
	Judge(ctx context.Context, req *JudgeRequest) (*JudgeResponse, error)
}

// serviceDesc describes the Judger service using raw grpc primitives. We
// avoid depending on protoc-generated stubs — the codec is already set to
// JSON and the method path matches proto/judger.proto.
var serviceDesc = grpc.ServiceDesc{
	ServiceName: "judger.Judger",
	HandlerType: (*Handler)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Judge",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(JudgeRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(Handler).Judge(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: MethodJudge}
				h := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(Handler).Judge(ctx, req.(*JudgeRequest))
				}
				return interceptor(ctx, in, info, h)
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "judger.proto",
}

// Register attaches Handler to the supplied *grpc.Server.
func Register(s *grpc.Server, h Handler) {
	s.RegisterService(&serviceDesc, h)
}

// MockSandbox is a deterministic pseudo-judger used when a real sandbox is
// unavailable. In production, replace this with an implementation that
// forwards to seccomp/cgroups isolated containers. The verdict depends on
// code length / language / problem id, so the same submission always yields
// the same result for a given problem.
type MockSandbox struct{}

func (MockSandbox) Judge(_ context.Context, req *JudgeRequest) (*JudgeResponse, error) {
	if strings.TrimSpace(req.Code) == "" {
		return &JudgeResponse{
			SubmissionID: req.SubmissionID,
			Status:       "Compile Error",
			RuntimeMS:    0,
			MemoryMB:     "0.0",
			MemoryKB:     0,
			CompileOut:   "empty source",
			ErrorMessage: "empty source",
		}, nil
	}
	if strings.Contains(req.Code, "segfault") {
		return &JudgeResponse{
			SubmissionID: req.SubmissionID,
			Status:       "Runtime Error",
			RuntimeMS:    42,
			MemoryMB:     "1.5",
			MemoryKB:     1536,
			ErrorMessage: "SIGSEGV",
		}, nil
	}

	h := sha1.Sum([]byte(fmt.Sprintf("%d|%d|%s", req.ProblemID, req.SubmissionID, req.Code)))
	seed := binary.BigEndian.Uint32(h[:4])
	status := "Accepted"
	switch seed % 10 {
	case 0, 1:
		status = "Wrong Answer"
	case 2:
		status = "Time Limit Exceeded"
	}
	if len(req.Code) > 8000 {
		status = "Time Limit Exceeded"
	}
	runtime := int32(10 + int(seed%300))
	memory := fmt.Sprintf("%.1f", 1.5+float64(seed%40)/10.0)
	memoryKB := int32((1500 + int(seed%40)*100))
	return &JudgeResponse{
		SubmissionID: req.SubmissionID,
		Status:       status,
		RuntimeMS:    runtime,
		MemoryMB:     memory,
		MemoryKB:     memoryKB,
		CaseResults: []CaseResult{
			{
				CaseNo:    1,
				Status:    status,
				RuntimeMS: runtime,
				MemoryKB:  memoryKB,
			},
		},
	}, nil
}
