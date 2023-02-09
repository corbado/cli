package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/corbado/cli/pkg/cli"
)

func TestLoginWithInvalidProjectID(t *testing.T) {
	stdout, stderr, err := cli.New(nil).ExecuteWithArgs("login", "--projectID=invalid", "--cliSecret=invalid")
	assert.NotNil(t, err)
	assert.Contains(t, stdout, "corbado login [flags]")
	assert.Contains(t, stderr, "Error: invalid projectID, must be of format pro-<number>")
}

func TestLoginWithInvalidCliSecret(t *testing.T) {
	tunnelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.Equal(t, "pro-1", username)
		assert.Equal(t, "invalid", password)
		assert.True(t, ok)

		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer tunnelServer.Close()

	consoleOutput := new(bytes.Buffer)

	stdout, stderr, err := cli.New(consoleOutput).ExecuteWithArgs(
		"login",
		"--projectID=pro-1",
		"--cliSecret=invalid",
		fmt.Sprintf("--tunnelAddress=ws%s", strings.TrimPrefix(tunnelServer.URL, "http")),
	)
	assert.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Contains(t, consoleOutput.String(), "failed (invalid credentials)")
}

func TestLoginSuccess(t *testing.T) {
	tunnelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.Equal(t, "pro-1", username)
		assert.Equal(t, "valid", password)
		assert.True(t, ok)

		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer c.Close()
	}))
	defer tunnelServer.Close()

	consoleOutput := new(bytes.Buffer)
	credentialFile := t.TempDir() + "/credentialFile"

	stdout, stderr, err := cli.New(consoleOutput).ExecuteWithArgs(
		"login",
		"--projectID=pro-1",
		"--cliSecret=valid",
		fmt.Sprintf("--credentialFile=%s", credentialFile),
		fmt.Sprintf("--tunnelAddress=ws%s", strings.TrimPrefix(tunnelServer.URL, "http")),
	)
	assert.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Contains(t, consoleOutput.String(), "success")

	credentialFileContent, err := os.ReadFile(credentialFile)
	assert.NoError(t, err)
	assert.Contains(t, string(credentialFileContent), "pro-1")
	assert.Contains(t, string(credentialFileContent), "valid")
}
