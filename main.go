package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "detect":
		buildpackYAMLParser := NewBuildpackYAMLParser()

		packit.Detect(Detect(buildpackYAMLParser))

	case "build":
		entryResolver := NewPlanEntryResolver()
		dependencyManager := postal.NewService(cargo.NewTransport())
		planRefinery := NewBuildPlanRefinery()
		clock := NewClock(time.Now)
		logEmitter := NewLogEmitter(os.Stdout)

		packit.Build(Build(entryResolver, dependencyManager, planRefinery, clock, logEmitter))

	default:
		panic(fmt.Sprintf("unknown command: %s", filepath.Base(os.Args[0])))
	}
}
