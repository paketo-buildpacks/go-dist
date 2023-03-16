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

func testEnvironmentVariableConfiguration(t *testing.T, context spec.G, it spec.S) {
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

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("env vars are used for configuration", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
				Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			})

			it("builds with the version from BP_GO_VERSION", func() {
				var (
					logs fmt.Stringer
					err  error
				)

				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(buildpack, buildPlanBuildpack).
					WithEnv(map[string]string{"BP_GO_VERSION": "1.20.*"}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithCommand("go run main.go").
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve(ContainSubstring("go1.20")).OnPort(8080))

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
					"  Resolving Go version",
					"    Candidate version sources (in priority order):",
					"      BP_GO_VERSION -> \"1.20.*\"",
					"      <unknown>     -> \"\"",
					"",
					MatchRegexp(`    Selected Go version \(using BP_GO_VERSION\): 1\.20\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Go 1\.20\.\d+`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				))
			})
		})

		context("the app contains a buildpack.yml", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(source, "buildpack.yml"), nil, os.ModePerm)).To(Succeed())
			})
			it("fails the build", func() {
				var err error

				_, logs, err := pack.Build.
					WithPullPolicy("never").
					WithBuildpacks(buildpack, buildPlanBuildpack).
					Execute(name, source)
				Expect(err).To(HaveOccurred(), logs.String)

				Expect(logs).To(ContainSubstring("working directory contains deprecated 'buildpack.yml'; use environment variables for configuration"))
				Expect(logs).NotTo(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
					"  Executing build process",
				))
			})

		})
	})
}
