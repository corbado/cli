package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/corbado/cli/pkg/ansi"
)

type CLI struct {
	out     io.Writer
	rootCmd *cobra.Command
}

const cliName = "corbado"

// New returns new CLI instance
func New(out io.Writer) *CLI {
	if out == nil {
		out = os.Stdout
	}

	c := &CLI{
		out: out,
	}

	c.defineCommands()

	return c
}

// Execute executes command
func (c *CLI) Execute() error {
	if err := c.rootCmd.Execute(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// ExecuteWithArgs executes command with given arguments
func (c *CLI) ExecuteWithArgs(args ...string) (string, string, error) {
	c.rootCmd.SetArgs(args)

	stdout := new(bytes.Buffer)
	c.rootCmd.SetOut(stdout)

	stderr := new(bytes.Buffer)
	c.rootCmd.SetErr(stderr)

	_, err := c.rootCmd.ExecuteC()

	return stdout.String(), stderr.String(), errors.WithStack(err)
}

func (c *CLI) defineCommands() {
	// Login
	loginCmd := &cobra.Command{
		Use:     "login",
		Example: cliName + " login",
		Short:   "Logs in with provided credentials",
		RunE:    c.handleLogin,
	}
	loginCmd.PersistentFlags().String("projectID", "", "ID of the project you want to login to")
	loginCmd.PersistentFlags().String("cliSecret", "", "CLI secret for the given project ID (can be found at https://app.corbado.com/app/settings/credentials/cli-secret)")
	loginCmd.PersistentFlags().String("tunnelAddress", "wss://tunnel1.corbado.com/v1", "Address of the Corbado tunnel server")
	loginCmd.PersistentFlags().String("credentialFile", "$HOME/.corbado", "Credentials file location")

	// Logout
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logs out (removes credentials file)",
		RunE:  c.handleLogout,
	}
	logoutCmd.PersistentFlags().Bool("force", false, "Forces logout (skips interaction)")
	logoutCmd.PersistentFlags().String("credentialFile", "$HOME/.corbado", "Credentials file location")

	// Subscribe
	subscribeCmd := &cobra.Command{
		Use:     "subscribe <localAddress>",
		Example: cliName + " subscribe http://localhost:8000",
		Short:   "Subscribes to webhook requests",
		RunE:    c.handleSubscribe,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("There must be only one argument and it must be your local address")
			}

			return nil
		},
	}

	subscribeCmd.PersistentFlags().String("tunnelAddress", "wss://tunnel1.corbado.com/v1", "Address of the Corbado tunnel server")
	subscribeCmd.PersistentFlags().String("projectID", "", "ID of the project you want to get webhook requests for")
	subscribeCmd.PersistentFlags().String("cliSecret", "", "CLI secret for the given project ID (can be found at https://app.corbado.com/app/settings/credentials/cli-secret)")
	subscribeCmd.PersistentFlags().String("credentialFile", "$HOME/.corbado", "Credentials file location")

	// Root
	c.rootCmd = &cobra.Command{Use: cliName}
	c.rootCmd.PersistentFlags().Bool("colors", true, "Defines if colors are used on output")
	c.rootCmd.AddCommand(loginCmd, logoutCmd, subscribeCmd)
}

func (c *CLI) getAnsi() (*ansi.Ansi, error) {
	useColors, err := c.rootCmd.PersistentFlags().GetBool("colors")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ansi.New(useColors, os.Stdout), nil
}

func (c *CLI) print(a ...any) {
	if _, err := fmt.Fprint(c.out, a...); err != nil {
		panic(err)
	}
}

func (c *CLI) println(a ...any) {
	if _, err := fmt.Fprintln(c.out, a...); err != nil {
		panic(err)
	}
}

func (c *CLI) printf(format string, a ...any) {
	if _, err := fmt.Fprintf(c.out, format, a...); err != nil {
		panic(err)
	}
}

func buildCredentialFilePath(credentialFile string) (string, error) {
	if credentialFile != "$HOME/.corbado" { //nolint:gosec
		// User overwrote flag, just return what he
		// has set
		return credentialFile, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.WithStack(err)
	}

	return fmt.Sprintf("%s/%s", homeDir, ".corbado"), nil
}
