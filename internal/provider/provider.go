package provider

import (
	"io"
	"net/http"

	"github.com/Tanrocode/layway/internal/schema"
)

type Provider interface {
	Endpoint() string
	Translate(req schema.ChatRequest) (io.Reader, error)
	Forward(body io.Reader) (*http.Response, error)
	ParseResponse(resp *http.Response) (schema.ChatResponse, error)
}
