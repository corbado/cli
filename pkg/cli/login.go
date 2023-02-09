package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/corbado/cli/pkg/ansi"
	"github.com/corbado/cli/pkg/tunnel"
)

const exitCommand = "exit"

func (c *CLI) handleLogin(cmd *cobra.Command, _ []string) error {
	ansi, err := c.getAnsi()
	if err != nil {
		return err
	}

	projectID, err := cmd.PersistentFlags().GetString("projectID")
	if err != nil {
		return errors.WithStack(err)
	}

	if projectID == "" {
		// No projectID given as flag, read interactively
		exit := false
		projectID, exit = c.readProjectID(ansi)
		if exit {
			return nil
		}
	} else if !c.validateProjectID(projectID) {
		return errors.New("invalid projectID, must be of format pro-<number>")
	}

	cliSecret, err := cmd.PersistentFlags().GetString("cliSecret")
	if err != nil {
		return errors.WithStack(err)
	}

	if cliSecret == "" {
		// No cliSecret given as flag, read interactively
		exit := false
		cliSecret, exit = c.readCliSecret(ansi)
		if exit {
			return nil
		}
	}

	tunnelAddress, err := cmd.PersistentFlags().GetString("tunnelAddress")
	if err != nil {
		return errors.WithStack(err)
	}

	credentialFile, err := cmd.PersistentFlags().GetString("credentialFile")
	if err != nil {
		return errors.WithStack(err)
	}

	tun := tunnel.New(ansi, tunnelAddress)

	c.printf("Authenticating to tunnel server (%s) ... ", tunnelAddress)
	if err := tun.Connect(projectID, cliSecret); err != nil {
		switch err {
		case tunnel.ErrUnauthorized:
			c.println(ansi.Red("failed (invalid credentials)!"))

			return nil

		case tunnel.ErrSessionExists:
			c.println(ansi.Red("failed (session already exists)!"))

			return nil

		case tunnel.ErrInternal:
			c.println(ansi.Red("failed (internal server error)!"))

			return nil
		}

		return err
	}
	c.println(ansi.Green("success!"))

	credentialFilePath, err := writeCredentialFile(credentialFile, projectID, cliSecret)
	if err != nil {
		return err
	}

	if err := tun.Stop(); err != nil {
		return err
	}

	c.println(ansi.Green(fmt.Sprintf("Successfully logged in by writing the credential file '%s' (use logout command to remove)!\n", credentialFilePath)))

	return nil
}

func (c *CLI) readProjectID(ansi *ansi.Ansi) (string, bool) {
	projectID := ""
	for {
		c.print("Please give us your projectID: ")
		_, _ = fmt.Scanln(&projectID)

		if projectID == "" {
			c.println(ansi.Red("Empty projectID (type exit to exit)"))
			continue
		}

		if projectID == exitCommand {
			return "", true
		}

		if !c.validateProjectID(projectID) {
			c.println(ansi.Red("Invalid projectID, must be of format pro-<number> (type exit to exit)"))
			continue
		}

		break
	}

	return projectID, false
}

func (c *CLI) readCliSecret(ansi *ansi.Ansi) (string, bool) {
	cliSecret := ""
	for {
		c.print("Please give us your CLI secret (can be found at https://app.corbado.com/app/settings/credentials/cli-secret): ")
		_, _ = fmt.Scanln(&cliSecret)

		if cliSecret == "" {
			c.println(ansi.Red("Empty CLI secret (type exit to exit)"))
			continue
		}

		if cliSecret == exitCommand {
			return "", true
		}

		break
	}

	return cliSecret, false
}

func writeCredentialFile(credentialFile string, projectID string, cliSecret string) (string, error) {
	credentialFilePath, err := buildCredentialFilePath(credentialFile)
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(credentialFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", errors.WithStack(err)
	}

	fileWriter := bufio.NewWriter(f)
	if _, err := fileWriter.WriteString(fmt.Sprintf("%s\n%s", projectID, cliSecret)); err != nil {
		return "", errors.WithStack(err)
	}

	if err := fileWriter.Flush(); err != nil {
		return "", errors.WithStack(err)
	}

	if err := f.Close(); err != nil {
		return "", errors.WithStack(err)
	}

	return credentialFilePath, nil
}
