package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Generalized for handle function
type Provider interface {
	Endpoint() string
	Forward(body io.Reader) (*http.Response, error)
}

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func (p *OpenAIProvider) Endpoint() string {
	return p.BaseURL + "/v1/chat/completions"
}

func (p *OpenAIProvider) Forward(body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, p.Endpoint(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	return http.DefaultClient.Do(req)
}

type AnthropicProvider struct {
	APIKey  string
	BaseURL string
}

func (p *AnthropicProvider) Endpoint() string {
	return p.BaseURL + "/v1/messages"
}

func (p *AnthropicProvider) Forward(body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, p.Endpoint(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return http.DefaultClient.Do(req)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	mux := http.NewServeMux()

	if openaiKey != "" {
		openai := &OpenAIProvider{APIKey: openaiKey, BaseURL: "https://api.openai.com"}
		mux.HandleFunc("POST /openai/v1/chat/completions", handleChatCompletions(openai))
	} else {
		log.Println("OPENAI_API_KEY not set, skipping OpenAI route")
	}

	if anthropicKey != "" {
		anthropic := &AnthropicProvider{APIKey: anthropicKey, BaseURL: "https://api.anthropic.com"}
		mux.HandleFunc("POST /anthropic/v1/messages", handleChatCompletions(anthropic))
	} else {
		log.Println("ANTHROPIC_API_KEY not set, skipping Anthropic route")
	}

	log.Println("gateway listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

const maxRetryAttempts = 3

func forwardWithRetry(provider Provider, bodyBytes []byte) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		resp, err = provider.Forward(bytes.NewReader(bodyBytes))

		retryable := err != nil || resp.StatusCode >= 500
		if !retryable {
			return resp, err
		}

		if attempt < maxRetryAttempts-1 {
			backoff := time.Duration(1<<attempt) * 200 * time.Millisecond
			log.Printf("attempt %d failed, retrying in %s", attempt+1, backoff)
			time.Sleep(backoff)
		}
	}

	return resp, err
}

func handleChatCompletions(provider Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := forwardWithRetry(provider, bodyBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("error copying response body: %v", err)
		}
	}
}
