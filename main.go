package main

import (
	"log"
	"os"

	"github.com/quantumsheep/nyaa-cli/cmd"
)

func main() {
	err := cmd.RootCmd.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
