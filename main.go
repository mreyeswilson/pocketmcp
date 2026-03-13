package main

import (
	"fmt"
	"os"

	"github.com/mreyeswilson/pocketmcp/cmd/pocketmcp"
)

func main() {
	if err := pocketmcp.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
