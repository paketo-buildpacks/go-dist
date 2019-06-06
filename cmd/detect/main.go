package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/cloudfoundry/go-cnb/golang"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	version := context.BuildPlan[golang.Dependency].Version
	if exists {
		version, err = helper.ReadBuildpackYamlVersion(buildpackYAMLPath, "golang")
		if err != nil {
			return detect.FailStatusCode, err
		}
	}

	return context.Pass(buildplan.BuildPlan{
		golang.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": false},
		}})
}
