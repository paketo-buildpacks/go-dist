package components

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/versionology"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func ConvertReleaseToDependency(release Release, platform cargo.ConfigTarget) ([]versionology.Dependency, error) {
	var archive, source ReleaseFile
	for _, file := range release.Files {
		if file.OS == platform.OS && file.Arch == platform.Arch {
			archive = file
		}
		if file.Kind == "source" {
			source = file
		}
	}

	if (archive == ReleaseFile{} || source == ReleaseFile{}) {
		return nil, fmt.Errorf("could not find release file for linux/amd64")
	}

	purl := GeneratePURL("go", release.Version, archive.SHA256, archive.URL)

	licenses := GetBsd3LicenseInformation()

	// Validate the artifact
	archiveResponse, err := http.Get(archive.URL)
	if err != nil {
		return nil, err
	}
	defer archiveResponse.Body.Close()

	vr := cargo.NewValidatedReader(archiveResponse.Body, fmt.Sprintf("sha256:%s", archive.SHA256))
	valid, err := vr.Valid()
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("the given checksum of the artifact does not match with downloaded artifact")
	}

	// Validate the source
	sourceResponse, err := http.Get(source.URL)
	if err != nil {
		return nil, err
	}
	defer sourceResponse.Body.Close()

	vr = cargo.NewValidatedReader(sourceResponse.Body, fmt.Sprintf("sha256:%s", source.SHA256))
	valid, err = vr.Valid()
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("the given checksum of the source does not match with downloaded source")
	}

	dep := cargo.ConfigMetadataDependency{
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
	}

	allStacksDependency, err := versionology.NewDependency(dep, "*")
	if err != nil {
		return nil, fmt.Errorf("could not get create * dependency: %w", err)
	}

	return []versionology.Dependency{allStacksDependency}, nil

}

func GenerateMetadataWithPlatform(versionFetcher versionology.VersionFetcher, platform retrieve.Platform) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()

	fetcher := NewFetcher()
	releases, err := fetcher.Get()
	if err != nil {
		return nil, err
	}

	var allStacksDependency versionology.Dependency
	for _, release := range releases {
		if strings.TrimPrefix(release.Version, "go") == version {
			return ConvertReleaseToDependency(release, cargo.ConfigTarget{
				OS:   platform.OS,
				Arch: platform.Arch,
			})
		}
	}

	return []versionology.Dependency{allStacksDependency}, nil

}
