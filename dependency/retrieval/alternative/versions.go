package alternative

import (
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func FindNewVersions(path string, versions []string) ([]string, error) {
	config, err := cargo.NewBuildpackParser().Parse(path)
	if err != nil {
		return nil, err
	}

	var newVersions []string
	for _, constraint := range config.Metadata.DependencyConstraints {
		svConstraint, err := semver.NewConstraint(constraint.Constraint)
		if err != nil {
			return nil, err
		}

		latestVersion := semver.MustParse("0.0.0")
		for _, dependency := range config.Metadata.Dependencies {
			svVersion := semver.MustParse(dependency.Version)
			if svConstraint.Check(svVersion) && svVersion.GreaterThan(latestVersion) {
				latestVersion = svVersion
			}
		}

		for _, version := range versions {
			svVersion := semver.MustParse(version)
			if svConstraint.Check(svVersion) && svVersion.GreaterThan(latestVersion) {
				newVersions = append(newVersions, version)
			}
		}
	}

	return newVersions, nil
}
