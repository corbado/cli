package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/corbado/cli/pkg/cli"
)

func TestLogoutSuccess(t *testing.T) {
	consoleOutput := new(bytes.Buffer)
	credentialFile := t.TempDir() + "/credentialFile"

	err := os.WriteFile(credentialFile, []byte("xxx"), 0600)
	assert.NoError(t, err)
	assert.FileExists(t, credentialFile)

	stdout, stderr, err := cli.New(consoleOutput).ExecuteWithArgs(
		"logout",
		"--force",
		fmt.Sprintf("--credentialFile=%s", credentialFile),
	)
	assert.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Contains(t, consoleOutput.String(), "Successfully logged out")
	assert.NoFileExists(t, credentialFile)
}
