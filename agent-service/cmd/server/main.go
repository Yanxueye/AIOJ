package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-service/internal/ai"
	"agent-service/internal/config"
	"agent-service/internal/handler"
	"agent-service/internal/judge"
	"agent-service/internal/rag"

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
		cfg.EmbeddingModel,
	)
	judgeClient := judge.NewClient(cfg.AIOJBackendURL)

	// Initialize RAG service using langchaingo
	ragService := rag.NewService()

	// Set up embedder using AI client
	if aiClient != nil {
		ragService.SetEmbedder(&rag.AIEmbedder{
			EmbeddingFunc: func(text string) ([]float64, error) {
				return aiClient.Embedding(text)
			},
		})
	}

	go func() {
		if err := ragService.LoadFromDirectory("oiwiki_docs"); err != nil {
			log.Printf("[rag] failed to load oiwiki_docs/: %v", err)
		}
	}()

	h := handler.New(aiClient, judgeClient, ragService)

	r := gin.Default()

	// Request body size limit (1MB)
	r.MaxMultipartMemory = 1 << 20

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

	// Body size limit middleware (1MB)
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	api := r.Group("/api/agent")
	{
		api.GET("/health", h.Health)
		api.GET("/rag-status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"initialized": ragService.IsInitialized(),
				"documentCount": ragService.DocumentCount(),
			})
		})
		api.POST("/hint", h.Hint)
		api.POST("/analyze", h.Analyze)
		api.POST("/generate-solution", h.GenerateSolution)
		api.POST("/chat", h.Chat)
		// Routes matching AIOJ backend AI client expectations
		api.POST("/code-diagnosis", h.CodeDiagnosis)
		api.POST("/knowledge-graph", h.KnowledgeGraph)
		api.POST("/solve", h.Solve)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	go func() {
		log.Printf("[agent-service] starting on %s (provider=%s)", cfg.HTTPAddr, cfg.AIProvider)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[agent-service] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("[agent-service] stopped")
}
