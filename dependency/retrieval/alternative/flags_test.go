package alternative_test

import (
	"testing"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/alternative"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFlags(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("parses the flags", func() {
		buildpackTOMLPath, outputPath, err := alternative.ParseFlags([]string{
			"--buildpack-toml-path", "some-buildpack-toml-path",
			"--output", "some-output-path",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(buildpackTOMLPath).To(Equal("some-buildpack-toml-path"))
		Expect(outputPath).To(Equal("some-output-path"))
	})

	context("when an unknown flag is provided", func() {
		it("returns an error", func() {
			_, _, err := alternative.ParseFlags([]string{"--unknown"})
			Expect(err).To(MatchError("flag provided but not defined: -unknown"))
		})
	})
}
