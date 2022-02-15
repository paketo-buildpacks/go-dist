package godist

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(name string, entries []packit.BuildpackPlanEntry, priorites []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func Build(entryResolver EntryResolver, dependencyManager DependencyManager, sbomGenerator SBOMGenerator, clock chronos.Clock, logs GoLogger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logs.Process("Resolving Go version")
		entry, entries := entryResolver.Resolve(GoDependency, context.Plan.Entries, Priorities)
		logs.Candidates(entries)

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = "default"
		}

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.SelectedDependency(entry, dependency, clock.Now())
		bom := dependencyManager.GenerateBillOfMaterials(dependency)

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			logs.WarnBuildpackYML(context.BuildpackInfo.Version)
		}

		goLayer, err := context.Layers.Get(GoLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := entryResolver.MergeLayerTypes(GoDependency, context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: bom}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: bom}
		}

		cachedSHA, ok := goLayer.Metadata[DependencySHAKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", goLayer.Path)
			logs.Break()

			goLayer.Launch, goLayer.Build, goLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{goLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logs.Process("Executing build process")

		goLayer, err = goLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		goLayer.Launch, goLayer.Build, goLayer.Cache = launch, build, build

		logs.Subprocess("Installing Go %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencyManager.Deliver(dependency, context.CNBPath, goLayer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logs.Action("Completed in %s", duration.Round(time.Millisecond))
		logs.Break()

		logs.Process("Generating SBOM for directory %s", goLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, context.WorkingDir)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.Action("Completed in %s", duration.Round(time.Millisecond))
		logs.Break()

		goLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		goLayer.Metadata = map[string]interface{}{
			DependencySHAKey: dependency.SHA256,
			"built_at":       clock.Now().Format(time.RFC3339Nano),
		}

		return packit.BuildResult{
			Layers: []packit.Layer{goLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
