package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Tanrocode/layway/internal/provider"
	"github.com/Tanrocode/layway/internal/schema"
)

const maxRetryAttempts = 3

func forwardWithRetry(p provider.Provider, req schema.ChatRequest) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		var body io.Reader
		body, err = p.Translate(req)
		if err != nil {
			return nil, err
		}

		resp, err = p.Forward(body)

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

func HandleChatCompletions(providers []provider.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schema.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		var lastErr error
		for _, p := range providers {
			resp, err := forwardWithRetry(p, req)
			if err != nil {
				lastErr = err
				log.Printf("provider %s failed, falling back: %v", p.Endpoint(), err)
				continue
			}
			if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("provider %s returned status %d", p.Endpoint(), resp.StatusCode)
				resp.Body.Close()
				log.Printf("provider %s failed, falling back: %v", p.Endpoint(), lastErr)
				continue
			}

			parsed, err := p.ParseResponse(resp)
			if err != nil {
				lastErr = err
				continue
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(parsed)
			return
		}

		http.Error(w, "all providers failed: "+lastErr.Error(), http.StatusBadGateway)
	}
}
