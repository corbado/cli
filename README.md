# Corbdo CLI

The Corbado CLI (command line interface) is a tool which you can run locally on your command line to interact with the Corbado platform.

See our [official documentation](https://docs.corbado.com/helpful-guides/corbado-cli) for more details.

## Prerequisites
-secret
For most commands you need a projectID and cliSecret to authenticate. You can find both in our [developer panel](https://app.corbado.com/app/settings/credentials/cli-secret). We offer multiple authentication methods which are explained in our [documentation](https://docs.corbado.com/helpful-guides/corbado-cli#authentication).

## Installation

We offer multiple installation methods, see our [documentation](https://docs.corbado.com/helpful-guides/corbado-cli#installation) for more details. Go install of course always works:

`go install github.com/corbado/cli/cmd/corbado@latest`

Or check out our pre-compiled binaries at our [release page](https://github.com/corbado/cli/releases/latest).

## Run

To use the Corbado CLI just run:

`corbado`

It will print a list of all commands. See our [documentation](https://docs.corbado.com/helpful-guides/corbado-cli#commands) for a detailed explanation for each one of them.
