package main

import (
	"fmt"
	"os"

	"github.com/corbado/cli/pkg/cli"
)

func main() {
	if err := cli.New(nil).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
