package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
)

var prereleaseRegexp = regexp.MustCompile(`(rc|beta)\d+`)

type ReleaseFetcher struct {
	releases []Release
	m        sync.Mutex
}

func NewReleaseFetcher() *ReleaseFetcher {
	return &ReleaseFetcher{}
}

func (rf *ReleaseFetcher) List() ([]Release, error) {
	if err := rf.load(); err != nil {
		return nil, err
	}

	return rf.releases, nil
}

func (rf *ReleaseFetcher) Get(version string) (Release, error) {
	if err := rf.load(); err != nil {
		return Release{}, err
	}

	for _, release := range rf.releases {
		if release.SemVer.Equal(semver.MustParse(version)) {
			return release, nil
		}
	}

	return Release{}, fmt.Errorf("failed to find version %q", version)
}

func (rf *ReleaseFetcher) load() error {
	rf.m.Lock()
	defer rf.m.Unlock()

	response, err := http.Get("https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&rf.releases)
	if err != nil {
		return err
	}

	for i := range rf.releases {
		version := strings.TrimPrefix(rf.releases[i].Version, "go")
		version = prereleaseRegexp.ReplaceAllStringFunc(version, func(s string) string {
			return fmt.Sprintf("-%s", s)
		})

		rf.releases[i].SemVer, err = semver.NewVersion(version)
		if err != nil {
			return err
		}
	}

	return nil
}
