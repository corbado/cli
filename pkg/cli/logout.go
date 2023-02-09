package cli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *CLI) handleLogout(cmd *cobra.Command, _ []string) error {
	ansi, err := c.getAnsi()
	if err != nil {
		return err
	}

	force, err := cmd.PersistentFlags().GetBool("force")
	if err != nil {
		return errors.WithStack(err)
	}

	if !force {
		choice := ""
		for {
			c.print("Are you sure you want to log out (remove credentials file)? [yes/no/exit]: ")
			_, _ = fmt.Scanln(&choice)

			if choice == "" {
				c.println(ansi.Red("Empty choice (type exit to exit)"))
				continue
			}

			if choice == "exit" {
				return nil
			}

			if choice != "yes" && choice != "no" {
				c.println(ansi.Red("Invalid input, only yes, no or exit are allowed"))
				continue
			}

			break
		}

		if choice == "no" {
			return nil
		}
	}

	credentialFile, err := cmd.PersistentFlags().GetString("credentialFile")
	if err != nil {
		return errors.WithStack(err)
	}

	credentialFilePath, err := buildCredentialFilePath(credentialFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(credentialFilePath); errors.Is(err, os.ErrNotExist) {
		c.printf("Credential file '%s' does not exist, doing nothing and exitting\n", credentialFilePath)

		return nil
	}

	if err := os.Remove(credentialFilePath); err != nil {
		return errors.WithStack(err)
	}

	c.println(ansi.Green(fmt.Sprintf("Successfully logged out by deleting the credential file '%s'!", credentialFilePath)))

	return nil
}
