package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/corbado/cli/pkg/tunnel"
)

func (c *CLI) handleSubscribe(cmd *cobra.Command, args []string) error {
	ansi, err := c.getAnsi()
	if err != nil {
		return err
	}

	localAddress := args[0]

	vldMsg := c.validateLocalAddress(localAddress)
	if vldMsg != "" {
		return errors.Errorf("Invalid localAddress: %s", vldMsg)
	}

	tunnelAddress, err := cmd.PersistentFlags().GetString("tunnelAddress")
	if err != nil {
		return errors.WithStack(err)
	}

	credentialFile, err := cmd.PersistentFlags().GetString("credentialFile")
	if err != nil {
		return errors.WithStack(err)
	}

	credentialFilePath, err := buildCredentialFilePath(credentialFile)
	if err != nil {
		return err
	}

	projectID, cliSecret, err := c.getCredentials(NewFlagCredentials(cmd), NewEnvCredentials(), NewFileCredentials(credentialFilePath))
	if err != nil {
		return err
	}

	if !c.validateProjectID(projectID) {
		return errors.New("Invalid projectID")
	}

	tun := tunnel.New(ansi, tunnelAddress)

	c.printf("Subscribing to tunnel server (%s) to get webhook requests for %s ... ", tunnelAddress, ansi.Bold(localAddress))
	if err := tun.Connect(projectID, cliSecret); err != nil {
		switch err {
		case tunnel.ErrUnauthorized:
			c.println(ansi.Bold(ansi.Red("failed (invalid credentials)!")))

			return nil

		case tunnel.ErrSessionExists:
			c.println(ansi.Bold(ansi.Red("failed (another CLI is already connected)!")))

			return nil

		case tunnel.ErrInternal:
			c.println(ansi.Bold(ansi.Red("failed (internal server error)!")))

			return nil
		}

		return err
	}
	c.println(ansi.Bold(ansi.Green("success!")))

	if err := tun.Start(localAddress); err != nil {
		if err == tunnel.ErrConnectionClosed {
			c.println(err.Error())

			return nil
		}

		return err
	}

	return nil
}
