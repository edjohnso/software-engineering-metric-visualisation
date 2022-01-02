package main

import (
	"os"
	"fmt"
	"github.com/edjohnso/software-engineering-metric-visualisation/pkg/webserver"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <port> <templates>\n", os.Args[0])
		os.Exit(1)
	}
	if err := webserver.Start(":" + os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
