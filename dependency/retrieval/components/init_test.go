package components_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("go", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Dependency", testDependency)
	suite("License", testLicense)
	suite("Output", testOutput)
	suite("Purl", testPurl)
	suite("Releases", testReleases)
	suite("Versions", testVersions)
	suite.Run(t)
}
