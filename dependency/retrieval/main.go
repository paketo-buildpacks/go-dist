package main

import (
	"log"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/versionology"

	"github.com/paketo-buildpacks/libdependency/retrieve"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
)

type GoMetadata struct {
	SemverVersion *semver.Version
}

func (goMetadata GoMetadata) Version() *semver.Version {
	return goMetadata.SemverVersion
}

func main() {
	retrieve.NewMetadataWithPlatforms("go", getAllVersions, components.GenerateMetadataWithPlatform)
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
