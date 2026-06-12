package main

import (
	"log"

	"agent-service/internal/ai"
	"agent-service/internal/config"
	"agent-service/internal/handler"
	"agent-service/internal/judge"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	aiClient := ai.NewClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.OpenAIModel,
		cfg.OllamaURL,
		cfg.OllamaModel,
		cfg.AIProvider,
	)
	judgeClient := judge.NewClient(cfg.AIOJBackendURL)
	h := handler.New(aiClient, judgeClient)

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/agent")
	{
		api.GET("/health", h.Health)
		api.POST("/hint", h.Hint)
		api.POST("/analyze", h.Analyze)
		api.POST("/generate-solution", h.GenerateSolution)
		api.POST("/chat", h.Chat)
		// Routes matching AIOJ backend AI client expectations
		api.POST("/code-diagnosis", h.CodeDiagnosis)
		api.POST("/knowledge-graph", h.KnowledgeGraph)
		api.POST("/solve", h.Solve)
	}

	log.Printf("[agent-service] starting on %s (provider=%s)", cfg.HTTPAddr, cfg.AIProvider)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
