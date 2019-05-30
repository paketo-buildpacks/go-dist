package golang

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/spf13/cast"
)

const Dependency = "go"

type Contributor struct {
	layer              layers.DependencyLayer
	buildContribution  bool
	launchContribution bool
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantLayer := context.BuildPlan[Dependency]
	if !wantLayer {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	version := plan.Version
	if version == "" {
		if version, err = context.Buildpack.DefaultVersion(Dependency); err != nil {
			return Contributor{}, false, err
		}
	}

	dep, err := deps.Best(Dependency, version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{layer: context.Layers.DependencyLayer(dep)}
	contributor.buildContribution = cast.ToBool(plan.Metadata["build"])
	contributor.launchContribution = cast.ToBool(plan.Metadata["launch"])

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		return helper.ExtractTarGz(artifact, layer.Root, 1)
	}, c.flags()...)
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
