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

	if openaiKey != "" {
		openai := &provider.OpenAIProvider{APIKey: openaiKey, BaseURL: "https://api.openai.com"}
		handler := gateway.WithLogging(gateway.WithRateLimit(rateLimiter, gateway.HandleChatCompletions(openai)))
		mux.HandleFunc("POST /openai/v1/chat/completions", handler)
	} else {
		log.Println("OPENAI_API_KEY not set, skipping OpenAI route")
	}

	if anthropicKey != "" {
		anthropic := &provider.AnthropicProvider{APIKey: anthropicKey, BaseURL: "https://api.anthropic.com"}
		handler := gateway.WithLogging(gateway.WithRateLimit(rateLimiter, gateway.HandleChatCompletions(anthropic)))
		mux.HandleFunc("POST /anthropic/v1/messages", handler)
	} else {
		log.Println("ANTHROPIC_API_KEY not set, skipping Anthropic route")
	}

	log.Println("gateway listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
