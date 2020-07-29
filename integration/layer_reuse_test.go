package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}
	)

	it.Before(func() {
		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

	})

	context("when an app is rebuilt and does not change", func() {
		var (
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

		it("reuses a layer from a previous build", func() {
			firstImage, logs, err := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using <unknown>\): 1\.14\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.14\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			firstContainer, err := docker.Container.Run.WithCommand("go run main.go").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using <unknown>\): 1\.14\.\d+`),
				"",
				fmt.Sprintf("  Reusing cached layer /layers/%s/go", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))

			secondContainer, err := docker.Container.Run.WithCommand("go run main.go").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("go1.14"))

			Expect(secondImage.Buildpacks[0].Layers["go"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["go"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		var (
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
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("rebuilds the layer", func() {
			firstImage, logs, err := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.14.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using buildpack.yml\): 1\.14\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.14\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			firstContainer, err := docker.Container.Run.WithCommand("go run main.go").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			err = ioutil.WriteFile(filepath.Join(source, "buildpack.yml"), []byte(`---
go:
  version: 1.13.*
`), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Second pack build
			secondImage, logs, err := pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.13.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using buildpack.yml\): 1\.13\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.13\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			secondContainer, err := docker.Container.Run.WithCommand("go run main.go").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("go1.13"))

			Expect(secondImage.Buildpacks[0].Layers["go"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["go"].Metadata["built_at"]))
		})
	})
}
