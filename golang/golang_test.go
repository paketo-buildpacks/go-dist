package golang_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/cloudfoundry/go-cnb/golang"

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
			f.AddPlan(buildpackplan.Plan{Name: golang.Dependency})

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
		it("contributes golang to the build and launch layer but not cache layer when included in the build plan", func() {
			f.AddPlan(buildpackplan.Plan{
				Name:     golang.Dependency,
				Metadata: buildpackplan.Metadata{"build": true, "launch": false},
			})

			golangContributor, _, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(golangContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(golang.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
		})

		it("installs the golang dependency", func() {
			f.AddPlan(buildpackplan.Plan{
				Name: golang.Dependency,
			})

			golangContributor, _, err := golang.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(golangContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(golang.Dependency)
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})
	})
}
