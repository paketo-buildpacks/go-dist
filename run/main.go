package main

import (
	"os"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())

	packit.Run(
		godist.Detect(),
		godist.Build(entryResolver, dependencyManager, Generator{}, chronos.DefaultClock, logEmitter),
	)
}
