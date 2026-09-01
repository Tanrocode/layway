package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/Tanrocode/layway/internal/gateway"
	"github.com/Tanrocode/layway/internal/provider"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	rateLimiter := gateway.NewRateLimiter(10, time.Minute) // 10 requests per client per minute

	mux := http.NewServeMux()

	var allProviders []provider.Provider

	if openaiKey != "" {
		openai := &provider.OpenAIProvider{APIKey: openaiKey, BaseURL: "https://api.openai.com"}
		allProviders = append(allProviders, openai)
		handler := gateway.WithLogging(gateway.WithRateLimit(rateLimiter, gateway.HandleChatCompletions([]provider.Provider{openai})))
		mux.HandleFunc("POST /openai/v1/chat/completions", handler)
	} else {
		log.Println("OPENAI_API_KEY not set, skipping OpenAI route")
	}

	if anthropicKey != "" {
		anthropic := &provider.AnthropicProvider{APIKey: anthropicKey, BaseURL: "https://api.anthropic.com"}
		allProviders = append(allProviders, anthropic)
		handler := gateway.WithLogging(gateway.WithRateLimit(rateLimiter, gateway.HandleChatCompletions([]provider.Provider{anthropic})))
		mux.HandleFunc("POST /anthropic/v1/messages", handler)
	} else {
		log.Println("ANTHROPIC_API_KEY not set, skipping Anthropic route")
	}

	if len(allProviders) > 0 {
		handler := gateway.WithLogging(gateway.WithRateLimit(rateLimiter, gateway.HandleChatCompletions(allProviders)))
		mux.HandleFunc("POST /v1/chat/completions", handler)
	}

	log.Println("gateway listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
