package alternative_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitAlternative(t *testing.T) {
	suite := spec.New("dependency/retrieval/alternative", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Flags", testFlags)
	suite("Output", testOutput)
	suite("Versions", testVersions)
	suite.Run(t)
}
