package components

import (
	"errors"
	"net/http"
	"os"
	"sort"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/paketo-buildpacks/packit/v2/vacation"
)

func GetBsd3LicenseInformation() []interface{} {
	return []interface{}{
		"BSD-3-Clause",
	}
}

func GenerateLicenseInformation(url string) ([]interface{}, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = vacation.NewArchive(response.Body).StripComponents(1).Decompress(dir)
	if err != nil {
		return nil, err
	}

	f, err := filer.FromDirectory(dir)
	if err != nil {
		return nil, err
	}

	ls, err := licensedb.Detect(f)
	if err != nil {
		if errors.Is(err, licensedb.ErrNoLicenseFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var licenseIDs []string
	for license := range ls {
		licenseIDs = append(licenseIDs, license)
	}

	sort.Strings(licenseIDs)

	var licenseIDsAsInterface []interface{}
	for _, licenseID := range licenseIDs {
		licenseIDsAsInterface = append(licenseIDsAsInterface, licenseID)
	}

	return licenseIDsAsInterface, nil
}
