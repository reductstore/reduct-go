package httpclient

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reductstore/reduct-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTLSHandshakeErrorIsConnectionError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.3")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(Option{
		BaseURL: server.URL,
		Timeout: defaultTimeout,
	})

	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	assert.Nil(t, resp)
	require.Error(t, err)

	var apiErr *model.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, ConnectionError, apiErr.Status)
	assert.NotEqual(t, URLParseError, apiErr.Status)
	assert.NotContains(t, apiErr.Message, "invalid url")
}

func TestNewHTTPClientWithCustomCA(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.3")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	caPath := writeTestCACert(t, server.Certificate())
	client := NewHTTPClient(Option{
		BaseURL:    server.URL,
		Timeout:    defaultTimeout,
		CACertPath: caPath,
	})

	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewHTTPClientWithInsecureSkipVerify(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Reduct-API", "v1.3")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(Option{
		BaseURL:            server.URL,
		Timeout:            defaultTimeout,
		InsecureSkipVerify: true,
	})

	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewHTTPClientWithInvalidCACertPath(t *testing.T) {
	client := NewHTTPClient(Option{
		BaseURL:    "https://example.com",
		Timeout:    defaultTimeout,
		CACertPath: filepath.Join(t.TempDir(), "missing-ca.pem"),
	})

	_, err := client.NewRequest(http.MethodGet, "/test", http.NoBody)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing-ca.pem")
}

func writeTestCACert(t *testing.T, cert *x509.Certificate) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test-ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	require.NoError(t, os.WriteFile(path, pemBytes, 0o600))
	return path
}
