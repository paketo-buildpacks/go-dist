package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYAMLParser := NewBuildpackYAMLParser()
	entryResolver := NewPlanEntryResolver()
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := NewBuildPlanRefinery()
	logEmitter := NewLogEmitter(os.Stdout)

	packit.Run(
		Detect(buildpackYAMLParser),
		Build(entryResolver, dependencyManager, planRefinery, chronos.DefaultClock, logEmitter),
	)
}
