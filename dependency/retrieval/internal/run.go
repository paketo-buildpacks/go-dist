package internal

import (
	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/alternative"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func Run(args []string) error {
	buildpackTOMLPath, outputPath, err := alternative.ParseFlags(args)
	if err != nil {
		panic(err)
	}

	fetcher := NewReleaseFetcher()
	releases, err := fetcher.List()
	if err != nil {
		panic(err)
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.SemVer.String())
	}

	newVersions, err := alternative.FindNewVersions(buildpackTOMLPath, versions)
	if err != nil {
		panic(err)
	}

	var dependencies []cargo.ConfigMetadataDependency
	for _, version := range newVersions {
		release, err := fetcher.Get(version)
		if err != nil {
			panic(err)
		}

		dependency, err := release.Dependency()
		if err != nil {
			panic(err)
		}

		dependencies = append(dependencies, dependency)
	}

	err = alternative.WriteOutput(outputPath, dependencies, "banana")
	if err != nil {
		panic(err)
	}

	return nil
}
