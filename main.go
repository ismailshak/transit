package main

import (
	"os"

	"github.com/ismailshak/transit/internal/cli"
)

func main() {
	exit := cli.Run()
	os.Exit(exit)
}
