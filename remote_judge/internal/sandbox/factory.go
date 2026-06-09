package sandbox

import (
	"context"

	"remote_judge/internal/config"
)

// Build 根据运行时配置创建沙箱。始终使用 Docker。
func Build(cfg config.Config) Sandbox {
	SeccompProfileMode = cfg.SeccompProfile

	sb := &DockerCLISandbox{TransferMode: cfg.DockerTransfer}
	if cfg.PrewarmImages {
		PrewarmImages(context.Background(), sb)
	}

	var final Sandbox
	if cfg.EnableContainerPool {
		pooled := NewPooledSandbox(sb, cfg.ContainerPoolMaxSize)
		final = pooled
	} else {
		final = sb
	}

	return NewCircuitBreaker(final)
}
