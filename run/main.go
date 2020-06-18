package main

import (
	"os"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYAMLParser := godist.NewBuildpackYAMLParser()
	entryResolver := godist.NewPlanEntryResolver()
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := godist.NewBuildPlanRefinery()
	logEmitter := godist.NewLogEmitter(os.Stdout)

	packit.Run(
		godist.Detect(buildpackYAMLParser),
		godist.Build(entryResolver, dependencyManager, planRefinery, chronos.DefaultClock, logEmitter),
	)
}
