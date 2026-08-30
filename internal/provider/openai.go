package provider

import (
	"io"
	"net/http"
)

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
