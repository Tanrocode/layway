package provider

import (
	"io"
	"net/http"
)

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
