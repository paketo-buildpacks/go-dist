package godist

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type BuildPlanRefinery struct{}

func NewBuildPlanRefinery() BuildPlanRefinery {
	return BuildPlanRefinery{}
}

func (r BuildPlanRefinery) BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlanEntry {
	return packit.BuildpackPlanEntry{
		Name: dependency.ID,
		Metadata: map[string]interface{}{
			"version":  dependency.Version,
			"licenses": []string{},
			"name":     dependency.Name,
			"sha256":   dependency.SHA256,
			"stacks":   dependency.Stacks,
			"uri":      dependency.URI,
		},
	}
}
