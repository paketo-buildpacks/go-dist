package main_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("go-compiler", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("BuildPlanRefinery", testBuildPlanRefinery)
	suite("BuildpackYAMLParser", testBuildpackYAMLParser)
	suite("Clock", testClock)
	suite("Detect", testDetect)
	suite("LogEmitter", testLogEmitter)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite.Run(t)
}
