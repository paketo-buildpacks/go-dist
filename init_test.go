package godist_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("go-dist", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("BuildPlanRefinery", testBuildPlanRefinery)
	suite("BuildpackYAMLParser", testBuildpackYAMLParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
