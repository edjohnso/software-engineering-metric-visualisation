package main

import (
	"log"
	"github.com/edjohnso/software-engineering-metric-visualisation/pkg/webserver"
)

func main() {
	log.Fatal(webserver.Start())
}
