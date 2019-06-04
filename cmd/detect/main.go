package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"gopkg.in/yaml.v2"

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
		version, err = readBuildpackYamlVersion(buildpackYAMLPath)
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

func readBuildpackYamlVersion(buildpackYAMLPath string) (string, error) {
	buf, err := ioutil.ReadFile(buildpackYAMLPath)
	if err != nil {
		return "", err
	}

	config := struct {
		Golang struct {
			Version string `yaml:"version"`
		} `yaml:"golang"`
	}{}
	if err := yaml.Unmarshal(buf, &config); err != nil {
		return "", err
	}

	return config.Golang.Version, nil
}
