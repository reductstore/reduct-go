package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reduct-go/model"
	"time"
)

const (
	APIVersion = "v1"
)

type HTTPClient interface {
	Post(ctx context.Context, path string, requestBody, responseData any) error
	Put(ctx context.Context, path string, requestBody, responseData any) error
	Get(ctx context.Context, path string, responseData any) error
	Head(ctx context.Context, path string) error
	Delete(ctx context.Context, path string) error
}

type HttpClientOption struct {
	BaseUrl   string
	ApiToken  string
	Timeout   time.Duration
	VerifySSL bool
}

type httpClient struct {
	client   *http.Client
	apiToken string
	url      string
}

func NewHTTPClient(option HttpClientOption) HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: option.Timeout,
		},
		url:      fmt.Sprintf("%s/api/%s", option.BaseUrl, APIVersion),
		apiToken: option.ApiToken,
	}
}

func (c *httpClient) setClientHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}

func (c *httpClient) Put(ctx context.Context, path string, requestBody, responseData any) error {
	if c.client == nil {
		return &model.APIError{
			Message: "http client is not initialized",
		}
	}
	// Marshal the request body to JSON
	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return &model.APIError{
				Message:  err.Error(),
				Original: err,
			}
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = nil
	}
	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url+path, reqBody)
	if err != nil {
		return &model.APIError{
			Original: err,
			Message:  err.Error(),
		}
	}
	// set reques headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	reductError := resp.Header.Get("X-Reduct-Error")
	if err != nil {
		return &model.APIError{
			Message:  err.Error(),
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	if responseData != nil && len(bodyBytes) > 0 {
		// Unmarshal the response into the provided responseData interface
		err := json.Unmarshal(bodyBytes, responseData)
		if err != nil {
			return &model.APIError{
				Message:  reductError,
				Original: err,
				Status:   resp.StatusCode,
			}
		}
	}
	return nil
}
func (c *httpClient) Post(ctx context.Context, path string, requestBody, responseData any) error {
	if c.client == nil {
		return &model.APIError{
			Message: "http client is not initialized",
		}
	}
	// Marshal the request body to JSON
	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return &model.APIError{
				Message:  err.Error(),
				Original: err,
			}
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = nil
	}
	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+path, reqBody)
	if err != nil {
		return &model.APIError{
			Original: err,
			Message:  err.Error(),
		}
	}
	// set reques headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	reductError := resp.Header.Get("X-Reduct-Error")

	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode}
	}
	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode}
	}
	if responseData != nil && len(bodyBytes) > 0 {
		// Unmarshal the response into the provided responseData interface
		err := json.Unmarshal(bodyBytes, responseData)
		if err != nil {
			return &model.APIError{
				Message:  reductError,
				Original: err,
				Status:   resp.StatusCode,
			}
		}
	}
	return nil
}

func (c *httpClient) Get(ctx context.Context, path string, responseData any) error {
	if c.client == nil {
		return &model.APIError{Message: "http client is not initialized"}
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+path, nil)
	if err != nil {
		return &model.APIError{Original: err}
	}
	// set reques headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	reductError := resp.Header.Get("X-Reduct-Error")

	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	if responseData != nil && len(bodyBytes) > 0 {
		// Unmarshal the response into the provided responseData interface
		err := json.Unmarshal(bodyBytes, responseData)
		if err != nil {
			return &model.APIError{
				Message:  reductError,
				Original: err,
				Status:   resp.StatusCode,
			}
		}
	}
	return nil
}

func (c *httpClient) Head(ctx context.Context, path string) error {
	if c.client == nil {
		return &model.APIError{Message: "http client is not initialized"}
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.url+path, nil)
	if err != nil {
		return &model.APIError{Original: err}
	}
	// set reques headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	reductError := resp.Header.Get("X-Reduct-Error")
	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}

	return nil
}

func (c *httpClient) Delete(ctx context.Context, path string) error {
	if c.client == nil {
		return &model.APIError{Message: "http client is not initialized"}
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url+path, nil)
	if err != nil {
		return &model.APIError{Original: err}
	}
	// set reques headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the request
	resp, err := c.client.Do(req)
	reductError := resp.Header.Get("X-Reduct-Error")

	if err != nil {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close response body: %v", err)
		}
	}()

	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return &model.APIError{
			Message:  reductError,
			Original: err,
			Status:   resp.StatusCode,
		}
	}

	return nil
}
