package main

import (
	"os"

	"github.com/TerrorSquad/forge/internal/forge"
)

func main() {
	os.Exit(forge.Run(os.Args[1:]))
}
