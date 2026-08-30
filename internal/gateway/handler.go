package gateway

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Tanrocode/layway/internal/provider"
)

const maxRetryAttempts = 3

func forwardWithRetry(p provider.Provider, bodyBytes []byte) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		resp, err = p.Forward(bytes.NewReader(bodyBytes))

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

func HandleChatCompletions(p provider.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := forwardWithRetry(p, bodyBytes)
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
