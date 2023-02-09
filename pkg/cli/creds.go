package cli

import (
	"bufio"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var ErrMissingProjectID = errors.New("missing project ID")
var ErrMissingCliSecret = errors.New("missing CLI secret")

type CredentialGetter interface {
	Get() (string, string, error)
}

func (c *CLI) getCredentials(options ...CredentialGetter) (string, string, error) {
	var projectID, secret string
	var err error

	for _, option := range options {
		projectID, secret, err = option.Get()
		if err == nil {
			return projectID, secret, nil
		}
	}

	return "", "", err
}

type FlagCredentials struct {
	cmd *cobra.Command
}

func NewFlagCredentials(cmd *cobra.Command) *FlagCredentials {
	return &FlagCredentials{
		cmd: cmd,
	}
}

func (f *FlagCredentials) Get() (string, string, error) {
	projectID, err := f.cmd.PersistentFlags().GetString("projectID")
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	if projectID == "" {
		return "", "", ErrMissingProjectID
	}

	cliSecret, err := f.cmd.PersistentFlags().GetString("cliSecret")
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	if cliSecret == "" {
		return "", "", ErrMissingCliSecret
	}

	return projectID, cliSecret, nil
}

type EnvCredentials struct{}

func NewEnvCredentials() *EnvCredentials {
	return &EnvCredentials{}
}

func (e *EnvCredentials) Get() (string, string, error) {
	projectID := os.Getenv("CORBADO_PROJECT_ID")
	if projectID == "" {
		return "", "", ErrMissingProjectID
	}

	cliSecret := os.Getenv("CORBADO_CLI_SECRET")
	if cliSecret == "" {
		return "", "", ErrMissingCliSecret
	}

	return projectID, cliSecret, nil
}

type FileCredentials struct {
	name string
}

func NewFileCredentials(name string) *FileCredentials {
	return &FileCredentials{
		name: name,
	}
}

func (f *FileCredentials) Get() (string, string, error) {
	fileReader, err := os.Open(f.name)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	reader := bufio.NewReader(fileReader)
	projectIDBytes, _, err := reader.ReadLine()
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	projectID := string(projectIDBytes)

	if projectID == "" {
		return "", "", ErrMissingProjectID
	}

	cliSecretBytes, _, err := reader.ReadLine()
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	cliSecret := string(cliSecretBytes)

	if cliSecret == "" {
		return "", "", ErrMissingCliSecret
	}

	return projectID, cliSecret, nil
}
