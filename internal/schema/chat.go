package schema

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}
