package components

import (
	"net/url"

	"github.com/package-url/packageurl-go"
)

// GeneratePURL can be used to populate the `purl` field of dependency metadata
// PURL stands for package URL.
// https://github.com/package-url/purl-spec
func GeneratePURL(id, version, checksum, source string) string {
	purl := packageurl.NewPackageURL(
		packageurl.TypeGeneric,
		"",
		id,
		version,
		packageurl.QualifiersFromMap(map[string]string{
			"checksum":     checksum,
			"download_url": source,
		}),
		"",
	)

	// Unescape the path to remove the added `%2F` and other encodings added to
	// the URL by packageurl-go
	// If the unescaping fails, we should still return the path URL with the
	// encodings, packageurl-go has examples with both the encodings and without,
	// we prefer to avoid the encodings when we can for convenience.
	purlString, err := url.PathUnescape(purl.ToString())
	if err != nil {
		return purl.ToString()
	}

	return purlString
}
