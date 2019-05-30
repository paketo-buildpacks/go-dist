package golang_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/go-cnb/golang"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitGolang(t *testing.T) {
	spec.Run(t, "Golang", testGolang, spec.Report(report.Terminal{}))
}

func testGolang(t *testing.T, when spec.G, it spec.S) {
	var (
		f                 *test.BuildFactory
		stubGolangFixture = filepath.Join("testdata", "stub-golang.tar.gz")
	)

	it.Before(func() {
		RegisterTestingT(t)
		f = test.NewBuildFactory(t)
		f.AddDependency(golang.Dependency, stubGolangFixture)
	})

	when("node.NewContributor", func() {
		it("returns true if a build plan exists", func() {
			f.AddBuildPlan(golang.Dependency, buildplan.Dependency{})

			_, willContribute, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			_, willContribute, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("Contribute", func() {
		it("contributes golang to the build and cache layer but not cachen layer when included in the build plan", func() {
			f.AddBuildPlan(golang.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true, "launch": false},
			})

			golangContributor, _, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(golangContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(golang.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
		})

		it("installs the golang dependency", func() {
			f.AddBuildPlan(golang.Dependency, buildplan.Dependency{})

			golangContributor, _, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(golangContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(golang.Dependency)
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})

		it("uses the default version when a version is not requested", func() {
			f.AddDependencyWithVersion(golang.Dependency, "0.9", filepath.Join("testdata", "stub-golang-default.tar.gz"))
			f.SetDefaultVersion(golang.Dependency, "0.9")
			f.AddBuildPlan(golang.Dependency, buildplan.Dependency{})

			golangContributor, _, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(golangContributor.Contribute()).To(Succeed())
			layer := f.Build.Layers.Layer(golang.Dependency)
			Expect(layer).To(test.HaveLayerVersion("0.9"))

			Expect(filepath.Join(layer.Root, "version-0.9.txt")).To(BeARegularFile())
		})

		it("returns an error when unsupported version of golang is included in the build plan", func() {
			f.AddBuildPlan(golang.Dependency, buildplan.Dependency{
				Version:  "9000.0.0",
				Metadata: buildplan.Metadata{"launch": true},
			})

			_, shouldContribute, err := golang.NewContributor(f.Build)
			Expect(err).To(HaveOccurred())
			Expect(shouldContribute).To(BeFalse())
		})
	})
}
