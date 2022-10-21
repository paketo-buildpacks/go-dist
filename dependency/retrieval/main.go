package main

import (
	"flag"
	"log"
	"os"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func main() {
	var buildpackTOMLPath, outputPath string
	set := flag.NewFlagSet("", flag.ContinueOnError)
	set.StringVar(&buildpackTOMLPath, "buildpack-toml-path", "", "path to the buildpack.toml file")
	set.StringVar(&outputPath, "output", "", "path to the output file")
	err := set.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	fetcher := components.NewFetcher()
	releases, err := fetcher.Get()
	if err != nil {
		log.Fatal(err)
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.SemVer.String())
	}

	newVersions, err := components.FindNewVersions(buildpackTOMLPath, versions)
	if err != nil {
		log.Fatal(err)
	}

	var dependencies []cargo.ConfigMetadataDependency
	for _, version := range newVersions {
		for _, r := range releases {
			if r.SemVer.String() == version {
				dependency, err := components.ConvertReleaseToDependency(r)
				if err != nil {
					log.Fatal(err)
				}
				dependencies = append(dependencies, dependency)
			}
		}
	}

	err = components.WriteOutput(outputPath, dependencies, "")
	if err != nil {
		log.Fatal(err)
	}
}
