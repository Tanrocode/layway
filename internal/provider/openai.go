package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Tanrocode/layway/internal/schema"
)

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func (p *OpenAIProvider) Endpoint() string {
	return p.BaseURL + "/v1/chat/completions"
}

func (p *OpenAIProvider) Translate(req schema.ChatRequest) (io.Reader, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
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

func (p *OpenAIProvider) ParseResponse(resp *http.Response) (schema.ChatResponse, error) {
	defer resp.Body.Close()
	var out schema.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return schema.ChatResponse{}, err
	}
	return out, nil
}
