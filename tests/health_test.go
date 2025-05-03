package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReductStoreHealth(t *testing.T) {
	healthUrl := "http://127.0.0.1:8383/api/v1/alive"

	req, err := http.NewRequest(http.MethodHead, healthUrl, nil)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, resp.StatusCode, http.StatusOK)

}
