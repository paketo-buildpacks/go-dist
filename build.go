package godist

import (
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(name string, entries []packit.BuildpackPlanEntry, priorites []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch bool, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface PlanRefinery --output fakes/plan_refinery.go
type PlanRefinery interface {
	BillOfMaterials(postal.Dependency) packit.BuildpackPlanEntry
}

func Build(entryResolver EntryResolver, dependencies DependencyManager, planRefinery PlanRefinery, clock chronos.Clock, logs scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logs.Process("Resolving Go version")
		entry, entries := entryResolver.Resolve(GoDependency, context.Plan.Entries, Priorities)
		logs.Candidates(entries)

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = "default"
		}

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.SelectedDependency(entry, dependency, clock.Now())
		bom := planRefinery.BillOfMaterials(dependency)

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logs.Subprocess("WARNING: Setting the Go Dist version through buildpack.yml will be deprecated soon in Go Dist Buildpack v%s.", nextMajorVersion.String())
			logs.Subprocess("Please specify the version through the $BP_GO_VERSION environment variable instead. See README.md for more information.")
			logs.Break()
		}

		goLayer, err := context.Layers.Get(GoLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := goLayer.Metadata[DependencySHAKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", goLayer.Path)
			logs.Break()

			return packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{bom},
				},
				Layers: []packit.Layer{goLayer},
			}, nil
		}

		logs.Process("Executing build process")

		goLayer, err = goLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		goLayer.Launch, goLayer.Build = entryResolver.MergeLayerTypes(GoDependency, context.Plan.Entries)
		goLayer.Cache = goLayer.Build

		logs.Subprocess("Installing Go %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, goLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logs.Action("Completed in %s", duration.Round(time.Millisecond))
		logs.Break()

		goLayer.Metadata = map[string]interface{}{
			DependencySHAKey: dependency.SHA256,
			"built_at":       clock.Now().Format(time.RFC3339Nano),
		}

		return packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{bom},
			},
			Layers: []packit.Layer{goLayer},
		}, nil
	}
}
