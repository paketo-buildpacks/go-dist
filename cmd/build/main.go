package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/paketo-buildpacks/go-compiler/golang"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	context.Logger.Title(context.Buildpack)

	goContributor, willContribute, err := golang.NewContributor(context)
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := goContributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildpackplan.Plan{})
}
