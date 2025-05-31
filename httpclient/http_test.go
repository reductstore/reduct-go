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
			apiVersion:     "v1.3",
			expectedError:  false,
			responseStatus: http.StatusOK,
		},
		{
			name:           "missing API version",
			apiVersion:     "",
			expectedError:  true,
			responseStatus: http.StatusOK,
		},
		{
			name:           "incompatible API version",
			apiVersion:     "v2.5",
			expectedError:  true,
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

			resp, err := client.Do(req)
			if tt.expectedError {
				defer resp.Body.Close()
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.responseStatus, resp.StatusCode)
			}
		})
	}
}
