package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const defaultTimeout = 30 * time.Second

func TestAPIVersionCheck(t *testing.T) {
	tests := []struct {
		name           string
		apiVersion     string
		expectedError  bool
		errorMessage   string
		responseStatus int
	}{
		{
			name:           "valid API version",
			apiVersion:     "1.8.0",
			expectedError:  false,
			responseStatus: http.StatusOK,
		},
		{
			name:           "missing API version",
			apiVersion:     "",
			expectedError:  true,
			errorMessage:   "Server did not provide API version",
			responseStatus: http.StatusOK,
		},
		{
			name:           "incompatible major version",
			apiVersion:     "2.0.0",
			expectedError:  true,
			errorMessage:   "Incompatible server API version",
			responseStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.apiVersion != "" {
					w.Header().Set("X-Reduct-API", tt.apiVersion)
				}
				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			// Create client
			client := NewHTTPClient(Option{
				BaseURL: server.URL,
				Timeout: defaultTimeout,
			})

			// Make request
			req, err := http.NewRequest(http.MethodGet, server.URL+"/test", http.NoBody)
			assert.NoError(t, err)

			resp, err := client.Do(req) //nolint:bodyclose // we don't need to close the body here

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.responseStatus, resp.StatusCode)
			}
		})
	}
}
