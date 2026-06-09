package sandbox

import (
	"context"

	"remote_judge/internal/domain"
	"remote_judge/internal/logger"
)

// ImageEnsurer 是镜像预热所需的沙箱子集接口。
type ImageEnsurer interface {
	ensureImage(ctx context.Context, image string) error
}

// PrewarmImages 在后台预拉取所有已配置的判题镜像。
func PrewarmImages(ctx context.Context, ensurer ImageEnsurer) {
	go func() {
		for _, lang := range domain.SupportedLanguages {
			if lang.DockerImage == "" {
				continue
			}
			if err := ensurer.ensureImage(ctx, lang.DockerImage); err != nil {
				logger.Error("sandbox.prewarm", "", "image prewarm failed", map[string]any{
					"image": lang.DockerImage,
					"error": err.Error(),
				})
				continue
			}
			logger.Info("sandbox.prewarm", "", "image prewarmed", map[string]any{"image": lang.DockerImage})
		}
		logger.Info("sandbox.prewarm", "", "all judge images pre-pulled", nil)
	}()
}
