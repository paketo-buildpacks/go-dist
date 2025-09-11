package main

import (
	"log"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/versionology"

	"github.com/paketo-buildpacks/libdependency/buildpack_config"
	"github.com/paketo-buildpacks/libdependency/retrieve"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type GoMetadata struct {
	SemverVersion *semver.Version
}

func (goMetadata GoMetadata) Version() *semver.Version {
	return goMetadata.SemverVersion
}

func main() {
	id := "go"
	buildpackTomlPath, output := retrieve.FetchArgs()
	retrieve.Validate(buildpackTomlPath, output)

	config, err := buildpack_config.ParseBuildpackToml(buildpackTomlPath)
	if err != nil {
		panic(err)
	}

	// We set by default the targets to linux/amd64 if no targets are specified in the buildpack.toml
	if len(config.Targets) == 0 {
		config.Targets = []cargo.ConfigTarget{
			{
				OS:   "linux",
				Arch: "amd64",
			},
		}
	}

	newVersions, err := retrieve.GetNewVersionsForId(id, config, getAllVersions)
	if err != nil {
		panic(err)
	}

	fetcher := components.NewFetcher()
	releases, err := fetcher.Get()
	if err != nil {
		panic(err)
	}

	var dependencies []cargo.ConfigMetadataDependency

	for _, target := range config.Targets {
		platform := cargo.ConfigTarget{
			OS:   target.OS,
			Arch: target.Arch,
		}

		// dependencies = append(dependencies, retrieve.GenerateAllMetadataWithPlatform(newVersions, generateMetadata, platform)...)

		for _, version := range newVersions {
			for _, r := range releases {
				if strings.TrimPrefix(r.Version, "go") == version.Version().String() {

					convertedDependency, err := components.ConvertReleaseToDependency(r, platform)
					if err != nil {
						log.Fatal(err)
					}
					dependencies = append(dependencies, convertedDependency)
				}
			}
		}
	}

	err = components.WriteOutput(output, dependencies, "")
	if err != nil {
		log.Fatal(err)
	}
}

func getAllVersions() (versionology.VersionFetcherArray, error) {

	fetcher := components.NewFetcher()
	goReleases, err := fetcher.Get()
	if err != nil {
		log.Fatal(err)
	}

	var versions []versionology.VersionFetcher
	for _, release := range goReleases {
		removePrefix := strings.TrimPrefix(release.Version, "go")
		goVersion, _ := semver.NewVersion(removePrefix)

		versions = append(versions, GoMetadata{
			goVersion,
		})
	}

	return versions, nil
}
