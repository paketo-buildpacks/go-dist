package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
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
			image occam.Image

			container1 occam.Container
			container2 occam.Container
			container3 occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container1.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container3.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds with the defaults", func() {
			var (
				logs fmt.Stringer
				err  error
			)

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			// Ensure go is installed correctly
			container1, err = docker.Container.Run.
				WithCommand("go run main.go").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container1).Should(Serve(ContainSubstring("go1.16")).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using <unknown>\): 1\.16\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.16\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			// check that all expected SBOM files are present
			container2, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("ls -al /layers/sbom/launch/%s/go/",
					strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container2.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(And(
				ContainSubstring("sbom.cdx.json"),
				ContainSubstring("sbom.spdx.json"),
				ContainSubstring("sbom.syft.json"),
			))

			// check an SBOM file to make sure it has an entry for go
			container3, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("cat /layers/sbom/launch/%s/go/sbom.cdx.json",
					strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container3.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring(`"name": "Go"`))
		})
	})
}
