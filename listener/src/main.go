package main

import (
	"github.com/dappnode/validator-monitoring/listener/src/server"
)

func main() {
	s := server.NewApi("8080")
	s.Start()
}
