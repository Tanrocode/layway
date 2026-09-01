package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Tanrocode/layway/internal/schema"
)

type AnthropicProvider struct {
	APIKey  string
	BaseURL string
}

func (p *AnthropicProvider) Endpoint() string {
	return p.BaseURL + "/v1/messages"
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const defaultMaxTokens = 1024

func (p *AnthropicProvider) Translate(req schema.ChatRequest) (io.Reader, error) {
	areq := anthropicRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
	}
	if areq.MaxTokens == 0 {
		areq.MaxTokens = defaultMaxTokens
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			areq.System = m.Content
			continue
		}
		areq.Messages = append(areq.Messages, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	data, err := json.Marshal(areq)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
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

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Model   string                  `json:"model"`
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

func (p *AnthropicProvider) ParseResponse(resp *http.Response) (schema.ChatResponse, error) {
	defer resp.Body.Close()
	var aresp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&aresp); err != nil {
		return schema.ChatResponse{}, err
	}

	var text string
	for _, block := range aresp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return schema.ChatResponse{
		Model: aresp.Model,
		Choices: []schema.Choice{
			{Message: schema.Message{Role: aresp.Role, Content: text}},
		},
	}, nil
}
