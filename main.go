package main

import (
	"os"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYAMLParser := NewBuildpackYAMLParser()
	entryResolver := NewPlanEntryResolver()
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := NewBuildPlanRefinery()
	clock := NewClock(time.Now)
	logEmitter := NewLogEmitter(os.Stdout)

	packit.Run(
		Detect(buildpackYAMLParser),
		Build(entryResolver, dependencyManager, planRefinery, clock, logEmitter),
	)
}
