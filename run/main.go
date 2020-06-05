package main

import (
	"os"

	gocompiler "github.com/paketo-buildpacks/go-compiler"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYAMLParser := gocompiler.NewBuildpackYAMLParser()
	entryResolver := gocompiler.NewPlanEntryResolver()
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := gocompiler.NewBuildPlanRefinery()
	logEmitter := gocompiler.NewLogEmitter(os.Stdout)

	packit.Run(
		gocompiler.Detect(buildpackYAMLParser),
		gocompiler.Build(entryResolver, dependencyManager, planRefinery, chronos.DefaultClock, logEmitter),
	)
}
