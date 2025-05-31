package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/reductstore/reduct-go/model"
)

const (
	APIVersion = "v1"
)

var (
	InvalidRequest  = -6 // used for invalid requests.
	Interrupt       = -5 // used for interrupting a long-running task or query.
	URLParseError   = -4 // used for invalid url.
	ConnectionError = -3 // used for network errors.
	Timeout         = -2 // used for timeout errors.
	Unknown         = -1 // used for unknown errors.
)

type HTTPClient interface {
	Post(ctx context.Context, path string, requestBody, responseData any) error
	Put(ctx context.Context, path string, requestBody, responseData any) error
	Patch(ctx context.Context, path string, requestBody, responseData any) error
	Get(ctx context.Context, path string, responseData any) error
	Head(ctx context.Context, path string) error
	Delete(ctx context.Context, path string) error
	Do(req *http.Request) (*http.Response, error)
	NewRequest(method, path string, body io.Reader) (*http.Request, error)
	NewRequestWithContext(ctx context.Context, method, path string, body io.Reader) (*http.Request, error)
}

type Option struct {
	BaseURL   string
	APIToken  string
	Timeout   time.Duration
	VerifySSL bool
}

type httpClient struct {
	client   *http.Client
	apiToken string
	url      string
}

func NewHTTPClient(option Option) HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: option.Timeout,
		},
		url:      fmt.Sprintf("%s/api/%s", option.BaseURL, APIVersion),
		apiToken: option.APIToken,
	}
}

func (c *httpClient) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.url+path, body)
	if err != nil {
		return nil, err
	}
	c.setClientHeaders(req)
	return req, nil
}

func (c *httpClient) NewRequestWithContext(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.url+path, body)
	if err != nil {
		return nil, err
	}
	c.setClientHeaders(req)
	return req, nil
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
	resp, err := c.Do(req)
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
			Message:  resp.Status,
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

func (c *httpClient) Patch(ctx context.Context, path string, requestBody, responseData any) error {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url+path, reqBody)
	if err != nil {
		return &model.APIError{
			Original: err,
			Message:  err.Error(),
		}
	}
	resp, err := c.Do(req)

	if err != nil {
		return handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

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
	resp, err := c.Do(req)

	if err != nil {
		return handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

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
func handleHTTPError(err error) error {
	var opErr *net.OpError
	var urlErr *url.Error

	switch {
	case errors.As(err, &opErr):
		return &model.APIError{
			Message:  "network error",
			Original: err,
			Status:   ConnectionError,
		}
	case errors.As(err, &urlErr):
		return &model.APIError{
			Message:  "invalid url",
			Original: err,
			Status:   URLParseError,
		}
	case errors.Is(err, http.ErrServerClosed):
		return &model.APIError{
			Message:  "server closed",
			Original: err,
			Status:   ConnectionError,
		}
	case errors.Is(err, context.Canceled):
		return &model.APIError{
			Message:  "request canceled",
			Original: err,
			Status:   Interrupt,
		}
	case errors.Is(err, context.DeadlineExceeded):
		return &model.APIError{
			Message:  "request timed out",
			Original: err,
			Status:   Timeout,
		}
	default:
		return &model.APIError{
			Message:  err.Error(),
			Original: err,
			Status:   Unknown,
		}
	}
}

func (c *httpClient) Get(ctx context.Context, path string, responseData any) error {
	if c.client == nil {
		return &model.APIError{Message: "http client is not initialized"}
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+path, http.NoBody)
	if err != nil {
		return &model.APIError{Original: err}
	}
	resp, err := c.Do(req)

	if err != nil {
		return handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

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
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.url+path, http.NoBody)
	if err != nil {
		return &model.APIError{Original: err}
	}
	resp, err := c.Do(req)

	if err != nil {
		return handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

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
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url+path, http.NoBody)
	if err != nil {
		return &model.APIError{Original: err}
	}
	resp, err := c.Do(req)

	if err != nil {
		return handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

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

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	// set request headers
	c.setClientHeaders(req)
	// Create an HTTP client and perform the Do
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, handleHTTPError(err)
	}
	reductError := resp.Header.Get("X-Reduct-Error")

	// Check API version header
	apiVersion := resp.Header.Get("X-Reduct-API")
	if apiVersion == "" {
		return nil, &model.APIError{
			Status:  resp.StatusCode,
			Message: "Server did not provide API version",
		}
	}

	// Check API version compatibility
	if err := model.CheckServerAPIVersion(apiVersion, APIVersion); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return resp, model.APIError{
			Status:  resp.StatusCode,
			Message: reductError,
		}
	}
	return resp, nil
}
