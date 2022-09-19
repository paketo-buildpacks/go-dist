package alternative

import (
	"encoding/json"
	"os"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type OutputDependency struct {
	cargo.ConfigMetadataDependency
	Target string `json:"target"`
}

func WriteOutput(path string, dependencies []cargo.ConfigMetadataDependency, target string) error {
	var output []OutputDependency
	for _, dependency := range dependencies {
		output = append(output, OutputDependency{
			ConfigMetadataDependency: dependency,
			Target:                   target,
		})
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(output)
	if err != nil {
		return err
	}

	return nil
}
