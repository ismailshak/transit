package main

import (
	"os"

	"github.com/ismailshak/transit/cmd"
)

func main() {
	exit := cmd.Run()
	os.Exit(exit)
}
