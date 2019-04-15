package main

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/go-cnb/golang"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {

	var (
		f *test.BuildFactory
		stubNodeFixture = filepath.Join("testdata", "stub-golang.tar.gz")
	)

	it.Before(func() {
		RegisterTestingT(t)

		f = test.NewBuildFactory(t)
	})

	it("always passes", func() {
		code, err := runBuild(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(build.SuccessStatusCode))
	})

	it("adds the go layer if in the build plan", func() {
		f.AddDependency(golang.Layer, stubNodeFixture)
		f.AddBuildPlan(golang.Layer, buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true},
		})

		code, err := runBuild(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(build.SuccessStatusCode))
		goLayer := f.Build.Layers.Layer(golang.Layer)
		Expect(goLayer).To(test.HaveLayerMetadata(true, true, false))
	})
}
