package aidarpc2

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// WebClient establishes a connection to a web server and provides a method to send requests and retrieve responses.
type WebClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewWebClient creates a new WebClient with the given base URL.
func NewWebClient(baseURL string) *WebClient {
	return &WebClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// SendRequest sends an HTTP request to the server and returns the response body and error.
func (c *WebClient) SendRequest(body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	res, err := io.ReadAll(resp.Body)
	return res, errors.Join(err, resp.Body.Close())
}

func (c *WebClient) Close() error {
	// No persistent connections to close in this simple implementation
	return nil
}
