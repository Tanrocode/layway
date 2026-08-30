package provider

import (
	"io"
	"net/http"
)

type Provider interface {
	Endpoint() string
	Forward(body io.Reader) (*http.Response, error)
}
