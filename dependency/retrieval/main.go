package main

import (
	"log"
	"os"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/internal"
)

func main() {
	if err := internal.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
