package main

import (
	"os"

	"github.com/TerrorSquad/gobooster/internal/booster"
)

func main() {
	os.Exit(booster.Run(os.Args[1:]))
}
