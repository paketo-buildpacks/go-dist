package main

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type BuildpackYAMLParser struct{}

func NewBuildpackYAMLParser() BuildpackYAMLParser {
	return BuildpackYAMLParser{}
}

func (p BuildpackYAMLParser) ParseVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}

		return "", fmt.Errorf("failed to open buildpack.yml: %w", err)
	}

	var config struct {
		Go struct {
			Version string `yaml:"version"`
		} `yaml:"go"`
	}

	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return "", fmt.Errorf("failed to decode buildpack.yml: %w", err)
	}

	return config.Go.Version, nil
}
