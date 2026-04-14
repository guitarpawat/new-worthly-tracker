package main

import (
	"os"

	"github.com/guitarpawat/worthly-tracker/internal/app"
)

func main() {
	app.Run(os.Args[1:])
}
