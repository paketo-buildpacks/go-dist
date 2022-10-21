package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Release struct {
	SemVer *semver.Version

	Version string        `json:"version"`
	Stable  bool          `json:"stable"`
	Files   []ReleaseFile `json:"files"`
}

type ReleaseFile struct {
	URL string

	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
	Kind     string `json:"kind"`
}

type Fetcher struct {
	releasePage string
}

func NewFetcher() Fetcher {
	return Fetcher{
		releasePage: "https://go.dev/dl/?mode=json&include=all",
	}
}

func (f Fetcher) WithReleasePage(uri string) Fetcher {
	f.releasePage = uri
	return f
}

func (f Fetcher) Get() ([]Release, error) {
	response, err := http.Get(f.releasePage)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode < 300) {
		return nil, fmt.Errorf("received a non 200 status code from %s: status code %d received", f.releasePage, response.StatusCode)
	}

	var releasesPage []Release
	err = json.NewDecoder(response.Body).Decode(&releasesPage)
	if err != nil {
		return nil, err
	}

	var releases []Release
	for _, release := range releasesPage {
		if !release.Stable {
			continue
		}

		release.SemVer, err = semver.NewVersion(strings.TrimPrefix(release.Version, "go"))
		if err != nil {
			return nil, fmt.Errorf("%w: the following version string could not be parsed %q", err, release.Version)
		}

		for i := range release.Files {
			release.Files[i].URL = fmt.Sprintf("https://go.dev/dl/%s", release.Files[i].Filename)
		}

		releases = append(releases, release)
	}

	return releases, nil
}
