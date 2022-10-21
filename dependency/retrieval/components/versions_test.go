package components_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVersions(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect

		buildpackTOMLPath string
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.Write([]byte(`
[metadata]

  [[metadata.dependencies]]
    version = "1.2.0"

  [[metadata.dependencies]]
    version = "1.2.1"

  [[metadata.dependencies]]
    version = "2.3.4"

  [[metadata.dependencies]]
    version = "2.3.5"

  [[metadata.dependency-constraints]]
    constraint = "1.2.*"
    id = "test"
    patches = 2

  [[metadata.dependency-constraints]]
    constraint = "2.3.*"
    id = "test"
    patches = 2
		`))
		Expect(err).NotTo(HaveOccurred())

		buildpackTOMLPath = file.Name()

	})

	it.After(func() {
		Expect(os.RemoveAll(buildpackTOMLPath)).To(Succeed())
	})

	context("FindNewVersions", func() {
		it("returns a list of all new versions", func() {
			newVersions, err := components.FindNewVersions(buildpackTOMLPath, []string{"1.2.2", "1.2.3", "1.2.4", "2.3.4", "2.3.5", "2.3.6", "9.9.9"})
			Expect(err).NotTo(HaveOccurred())

			Expect(newVersions).To(Equal([]string{"1.2.4", "1.2.3", "2.3.6"}))
		})
	})

	context("failure cases", func() {
		context("the buildpack cannot be parsed", func() {
			it.Before(func() {
				Expect(os.RemoveAll(buildpackTOMLPath)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := components.FindNewVersions(buildpackTOMLPath, []string{"1.2.2", "1.2.3", "1.2.4", "2.3.4", "2.3.5", "2.3.6", "9.9.9"})
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("the constraints cannot be parsed", func() {
			it.Before(func() {
				Expect(os.WriteFile(buildpackTOMLPath, []byte(`
[metadata]

  [[metadata.dependency-constraints]]
    constraint = "not-valid-constraint"
    id = "test"
    patches = 2
		`), os.ModePerm))
			})

			it("returns an error", func() {
				_, err := components.FindNewVersions(buildpackTOMLPath, []string{"1.2.2", "1.2.3", "1.2.4", "2.3.4", "2.3.5", "2.3.6", "9.9.9"})
				Expect(err).To(MatchError(ContainSubstring("improper constraint: not-valid-constraint")))
			})
		})
	})
}
