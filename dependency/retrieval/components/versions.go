package components

import (
	"sort"

	"github.com/Masterminds/semver"
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

		var compatibleVersions []string
		for _, version := range versions {
			svVersion := semver.MustParse(version)
			if svConstraint.Check(svVersion) && svVersion.GreaterThan(latestVersion) {
				compatibleVersions = append(compatibleVersions, version)
			}
		}

		sort.Slice(compatibleVersions, func(i, j int) bool {
			jVersion := semver.MustParse(compatibleVersions[j])
			iVersion := semver.MustParse(compatibleVersions[i])
			return iVersion.GreaterThan(jVersion)
		})

		if constraint.Patches > len(compatibleVersions) {
			newVersions = append(newVersions, compatibleVersions...)
		} else {
			newVersions = append(newVersions, compatibleVersions[:constraint.Patches]...)
		}
	}

	return newVersions, nil
}
