package godist_test

import (
	"os"
	"testing"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYAMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		parser godist.BuildpackYAMLParser
	)

	it.Before(func() {
		parser = godist.NewBuildpackYAMLParser()
	})

	context("ParseVersion", func() {
		var path string

		it.Before(func() {
			file, err := os.CreateTemp("", "buildpack.yml")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.WriteString(`---
go:
  version: some-version
`)
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Close()).To(Succeed())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("parses the version from a buildpack.yml file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("some-version"))
		})

		context("when the buildpack.yml file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty version without erroring", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(Equal(""))
			})
		})

		context("failure cases", func() {
			context("when the buildpack.yml file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("failed to open buildpack.yml:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the buildpack.yml file is malformed", func() {
				it.Before(func() {
					Expect(os.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("failed to decode buildpack.yml:")))
					Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
				})
			})
		})
	})
}
