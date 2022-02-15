package godist

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Build         bool   `toml:"build"`
	Version       string `toml:"version"`
}

func Detect(buildpackYAMLParser VersionParser) packit.DetectFunc {
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

		version, err := buildpackYAMLParser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: GoDependency,
				Metadata: BuildPlanMetadata{
					VersionSource: "buildpack.yml",
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
