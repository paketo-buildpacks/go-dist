package golang

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Dependency = "go"

type Contributor struct {
	layer              layers.DependencyLayer
	buildContribution  bool
	launchContribution bool
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantLayer , err := context.Plans.GetShallowMerged(Dependency)
	if err != nil{
		return Contributor{}, false, err

	}
	if !wantLayer {
		return Contributor{}, false, nil
	}

	dep, err := context.Buildpack.RuntimeDependency(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{layer: context.Layers.DependencyLayer(dep)}
	contributor.buildContribution, _ = plan.Metadata["build"].(bool)
	contributor.launchContribution, _ = plan.Metadata["launch"].(bool)

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)
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
