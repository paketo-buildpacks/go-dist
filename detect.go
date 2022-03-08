package godist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Build         bool   `toml:"build"`
	Version       string `toml:"version"`
}

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		bpYML, err := fs.Exists(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to check for buildpack.yml: %w", err)
		}
		if bpYML {
			return packit.DetectResult{}, fmt.Errorf("working directory contains deprecated 'buildpack.yml'; use environment variables for configuration")
		}

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
