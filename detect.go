package godist

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
)

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Build         bool   `toml:"build"`
	Version       string `toml:"version"`
}

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		if version, ok := os.LookupEnv("BP_GO_VERSION"); ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: GoDependency,
				Metadata: BuildPlanMetadata{
					VersionSource: "BP_GO_VERSION",
					Version:       version,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: GoDependency},
				},
				Requires: requirements,
			},
		}, nil
	}
}
