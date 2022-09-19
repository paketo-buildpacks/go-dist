package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/anchore/packageurl-go"
	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/vacation"
)

type Release struct {
	SemVer *semver.Version

	Version string        `json:"version"`
	Stable  bool          `json:"stable"`
	Files   []ReleaseFile `json:"files"`
}

func (r Release) Dependency() (cargo.ConfigMetadataDependency, error) {
	var archive, source ReleaseFile
	for _, file := range r.Files {
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

	purl := packageurl.NewPackageURL(
		packageurl.TypeGeneric,
		"",
		"go",
		r.Version,
		packageurl.QualifiersFromMap(map[string]string{
			"checksum":     archive.Checksum(),
			"download_url": archive.URI(),
		}),
		"",
	).ToString()

	purl, err := url.PathUnescape(purl)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer os.RemoveAll(dir)

	response, err := http.Get(source.URI())
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer response.Body.Close()

	err = vacation.NewArchive(response.Body).StripComponents(1).Decompress(dir)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	f, err := filer.FromDirectory(dir)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	ls, err := licensedb.Detect(f)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	var licenses []interface{}
	for license := range ls {
		licenses = append(licenses, license)
	}

	return cargo.ConfigMetadataDependency{
		ID:      "go",
		Name:    "Go",
		Version: r.SemVer.String(),
		Stacks: []string{
			"io.buildpacks.stacks.bionic",
			"io.paketo.stacks.tiny",
			"io.buildpacks.stacks.jammy",
			"io.buildpacks.stacks.jammy.tiny",
		},
		URI:            archive.URI(),
		Checksum:       archive.Checksum(),
		Source:         source.URI(),
		SourceChecksum: source.Checksum(),
		CPE:            fmt.Sprintf("cpe:2.3:a:golang:go:%s:*:*:*:*:*:*:*", strings.TrimPrefix(r.Version, "go")),
		PURL:           purl,
		Licenses:       licenses,
	}, nil
}

type ReleaseFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
	Kind     string `json:"kind"`
}

func (rf ReleaseFile) URI() string {
	return fmt.Sprintf("https://go.dev/dl/%s", rf.Filename)
}

func (rf ReleaseFile) Checksum() string {
	return fmt.Sprintf("sha256:%s", rf.SHA256)
}
