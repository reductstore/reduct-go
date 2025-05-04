package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(ctx context.Context, url, method string, requestBody, responseData any) error
}

// ClientModifier is a function that modifies the request
// it is used to add custom headers, or other modifications to the request
// before it is sent to the server
type ClientModifier func(req *http.Request)

type httpClient struct {
	client    *http.Client
	modifiers []ClientModifier
}

func NewHTTPClient(timeout time.Duration, modifiers ...ClientModifier) HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: timeout,
		},
		modifiers: modifiers,
	}
}

func (c *httpClient) Do(ctx context.Context, url, method string, requestBody, responseData any) error {
	if c.client == nil {
		return fmt.Errorf("http client is not initialized")
	}
	// Marshal the request body to JSON
	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = nil
	}
	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	for _, modifier := range c.modifiers {
		modifier(req)
	}
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if responseData != nil && len(bodyBytes) > 0 {
		// Unmarshal the response into the provided responseData interface
		return json.Unmarshal(bodyBytes, responseData)
	}
	return nil
}
