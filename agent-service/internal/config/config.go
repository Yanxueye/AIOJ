package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr string

	// Primary AI: OpenAI-compatible API (e.g., MIMO)
	AIProvider   string // "openai" or "ollama"
	OpenAIAPIKey string
	OpenAIBaseURL string
	OpenAIModel  string

	// Fallback AI: Local Ollama
	OllamaURL   string
	OllamaModel string

	// Judge service
	JudgeGRPCAddr  string
	AIOJBackendURL string
}

func Load() Config {
	// Load .env file if present
	loadEnvFile(".env")
	loadEnvFile("agent-service/.env")

	return Config{
		HTTPAddr:       getEnv("AGENT_HTTP_ADDR", ":8090"),
		AIProvider:     getEnv("AI_PROVIDER", "openai"),
		OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:  getEnv("OPENAI_BASE_URL", "https://token-plan-sgp.xiaomimimo.com/v1"),
		OpenAIModel:    getEnv("OPENAI_MODEL", "mimo-v2.5-pro"),
		OllamaURL:      getEnv("OLLAMA_URL", "http://127.0.0.1:11434"),
		OllamaModel:    getEnv("OLLAMA_MODEL", "qwen2.5-coder:7b"),
		JudgeGRPCAddr:  getEnv("JUDGE_GRPC_ADDR", "127.0.0.1:9090"),
		AIOJBackendURL: getEnv("AIOJ_BACKEND_URL", "http://127.0.0.1:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadEnvFile reads a .env file and sets environment variables (without overriding existing ones)
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Don't override existing env vars
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
