package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/paketo-buildpacks/go-compiler/golang"

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

	version := ""
	if exists {
		bpYml := &BuildpackYaml{}
		err = helper.ReadBuildpackYaml(buildpackYAMLPath, bpYml)
		if err != nil {
			return detect.FailStatusCode, err
		}
		version = bpYml.Go.Version
	}

	return context.Pass(buildplan.Plan{
		Provides: []buildplan.Provided{{Name: golang.Dependency}},
		Requires: []buildplan.Required{{
			Name:     golang.Dependency,
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": false},
		}},
	})
}

type BuildpackYaml struct {
	Go struct {
		Version string `yaml:"version"`
	} `yaml:"go"`
}
