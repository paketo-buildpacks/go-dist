package main

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/paketo-buildpacks/go-compiler/golang"

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
		f               *test.BuildFactory
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
		f.AddDependency(golang.Dependency, stubNodeFixture)
		f.AddPlan(buildpackplan.Plan{
			Name:     golang.Dependency,
			Metadata: buildpackplan.Metadata{"build": true},
		})

		code, err := runBuild(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(build.SuccessStatusCode))
		goLayer := f.Build.Layers.Layer(golang.Dependency)
		Expect(goLayer).To(test.HaveLayerMetadata(true, true, false))
	})
}
