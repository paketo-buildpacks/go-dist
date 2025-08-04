package components

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

func ConvertReleaseToDependencies(release Release, platforms []Platform) ([]cargo.ConfigMetadataDependency, error) {
	// For backward compatibility if platform is empty
	if len(platforms) == 0 {
		dependency, err := ConvertReleaseToDependency(release)
		if err != nil {
			return []cargo.ConfigMetadataDependency{}, err
		}
		return []cargo.ConfigMetadataDependency{dependency}, nil
	}

	var source ReleaseFile
	for _, file := range release.Files {
		if file.Kind == "source" {
			source = file
			break
		}
	}

	licenses, err := GenerateLicenseInformation(source.URL)
	if err != nil {
		return []cargo.ConfigMetadataDependency{}, err
	}

	var dependencies []cargo.ConfigMetadataDependency
	for _, platform := range platforms {
		var archive ReleaseFile
		for _, file := range release.Files {
			if file.OS == platform.OS && file.Arch == platform.Arch {
				archive = file
				break
			}
		}

		if (archive == ReleaseFile{} || source == ReleaseFile{}) {
			return []cargo.ConfigMetadataDependency{}, fmt.Errorf("could not find release file for %s/%s", platform.OS, platform.Arch)
		}

		purl := GeneratePURL("go", release.Version, archive.SHA256, archive.URL)

		// Validate the artifact
		archiveResponse, err := http.Get(archive.URL)
		if err != nil {
			return []cargo.ConfigMetadataDependency{}, err
		}
		defer archiveResponse.Body.Close()

		vr := cargo.NewValidatedReader(archiveResponse.Body, fmt.Sprintf("sha256:%s", archive.SHA256))
		valid, err := vr.Valid()
		if err != nil {
			return []cargo.ConfigMetadataDependency{}, err
		}

		if !valid {
			return []cargo.ConfigMetadataDependency{}, fmt.Errorf("the given checksum of the artifact does not match with downloaded artifact")
		}

		// Validate the source
		sourceResponse, err := http.Get(source.URL)
		if err != nil {
			return []cargo.ConfigMetadataDependency{}, err
		}
		defer sourceResponse.Body.Close()

		vr = cargo.NewValidatedReader(sourceResponse.Body, fmt.Sprintf("sha256:%s", source.SHA256))
		valid, err = vr.Valid()
		if err != nil {
			return []cargo.ConfigMetadataDependency{}, err
		}

		if !valid {
			return []cargo.ConfigMetadataDependency{}, fmt.Errorf("the given checksum of the source does not match with downloaded source")
		}

		dependencies = append(dependencies, cargo.ConfigMetadataDependency{
			ID:              "go",
			Name:            "Go",
			Version:         release.SemVer.String(),
			Stacks:          []string{"*"},
			URI:             archive.URL,
			Checksum:        fmt.Sprintf("sha256:%s", archive.SHA256),
			Source:          source.URL,
			SourceChecksum:  fmt.Sprintf("sha256:%s", source.SHA256),
			StripComponents: 1,
			CPE:             fmt.Sprintf("cpe:2.3:a:golang:go:%s:*:*:*:*:*:*:*", strings.TrimPrefix(release.Version, "go")),
			PURL:            purl,
			Licenses:        licenses,
			OS:              platform.OS,
			Arch:            platform.Arch,
		})
	}

	return dependencies, nil
}

func ConvertReleaseToDependency(release Release) (cargo.ConfigMetadataDependency, error) {
	var archive, source ReleaseFile
	for _, file := range release.Files {
		if file.OS == "linux" && file.Arch == "amd64" {
			archive = file
		}
		if file.Kind == "source" {
			source = file
		}
	}

	if (archive == ReleaseFile{} || source == ReleaseFile{}) {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("could not find release file for linux/amd64")
	}

	purl := GeneratePURL("go", release.Version, archive.SHA256, archive.URL)

	licenses, err := GenerateLicenseInformation(source.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	// Validate the artifact
	archiveResponse, err := http.Get(archive.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer archiveResponse.Body.Close()

	vr := cargo.NewValidatedReader(archiveResponse.Body, fmt.Sprintf("sha256:%s", archive.SHA256))
	valid, err := vr.Valid()
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	if !valid {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("the given checksum of the artifact does not match with downloaded artifact")
	}

	// Validate the source
	sourceResponse, err := http.Get(source.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer sourceResponse.Body.Close()

	vr = cargo.NewValidatedReader(sourceResponse.Body, fmt.Sprintf("sha256:%s", source.SHA256))
	valid, err = vr.Valid()
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	if !valid {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("the given checksum of the source does not match with downloaded source")
	}

	return cargo.ConfigMetadataDependency{
		ID:              "go",
		Name:            "Go",
		Version:         release.SemVer.String(),
		Stacks:          []string{"*"},
		URI:             archive.URL,
		Checksum:        fmt.Sprintf("sha256:%s", archive.SHA256),
		Source:          source.URL,
		SourceChecksum:  fmt.Sprintf("sha256:%s", source.SHA256),
		StripComponents: 1,
		CPE:             fmt.Sprintf("cpe:2.3:a:golang:go:%s:*:*:*:*:*:*:*", strings.TrimPrefix(release.Version, "go")),
		PURL:            purl,
		Licenses:        licenses,
	}, nil
}
