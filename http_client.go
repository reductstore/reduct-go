package reductgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DoRequest(ctx context.Context, url, method string, modifier func(req *http.Request) error, requestBody, responseData any) error {
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
	if modifier != nil {
		if err := modifier(req); err != nil {
			return err
		}
	}
	// Create an HTTP client and perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
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

	// Unmarshal the response into the provided responseData interface
	return json.Unmarshal(bodyBytes, responseData)
}
