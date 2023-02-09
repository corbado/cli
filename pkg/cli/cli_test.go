package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/corbado/cli/pkg/cli"
)

func TestEmptyCall(t *testing.T) {
	stdout, stderr, err := cli.New(nil).ExecuteWithArgs()
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Available Commands")
	assert.Empty(t, stderr)
}

func TestWithHelpFlag(t *testing.T) {
	stdout, stderr, err := cli.New(nil).ExecuteWithArgs("--help")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Available Commands")
	assert.Empty(t, stderr)
}

func TestWithUnknownFlag(t *testing.T) {
	stdout, stderr, err := cli.New(nil).ExecuteWithArgs("--unknown")
	assert.NotNil(t, err)
	assert.Contains(t, stdout, "Available Commands")
	assert.Contains(t, stderr, "unknown flag: --unknown")
}

func TestWithUnknownCommand(t *testing.T) {
	stdout, stderr, err := cli.New(nil).ExecuteWithArgs("unknown")
	assert.NotNil(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "unknown command \"unknown\"")
}
