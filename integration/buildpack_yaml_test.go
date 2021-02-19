package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackYAML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "buildpack_yaml_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds with the settings in buildpack.yml", func() {
			var (
				logs fmt.Stringer
				err  error
			)

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand("go run main.go").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve(ContainSubstring("go1.15")).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.15.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using buildpack\.yml\): 1\.15\.\d+`),
				"",
				"    WARNING: Setting the Go Dist version through buildpack.yml will be deprecated soon in Go Dist Buildpack v1.0.0.",
				"    Please specify the version through the $BP_GO_VERSION environment variable instead. See README.md for more information.",
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.15\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))
		})
	})
}
