package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
		name         string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

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

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			firstImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-buildpacks/go-compiler"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("go"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Go Compiler Buildpack %s", buildpackVersion),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using <unknown>\): 1\.13\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.13\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("go run main.go").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-buildpacks/go-compiler"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Go Compiler Buildpack %s", buildpackVersion),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using <unknown>\): 1\.13\.\d+`),
				"",
				"  Reusing cached layer /layers/paketo-buildpacks_go-compiler/go",
			))

			secondContainer, err = docker.Container.Run.WithCommand("go run main.go").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("go1.13"))

			Expect(secondImage.Buildpacks[0].Layers["go"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["go"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			firstImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "buildpack_yaml_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-buildpacks/go-compiler"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("go"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Go Compiler Buildpack %s", buildpackVersion),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.14.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using buildpack.yml\): 1\.14\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.14\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("go run main.go").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "different_version_buildpack_yaml_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-buildpacks/go-compiler"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("go"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Go Compiler Buildpack %s", buildpackVersion),
				"  Resolving Go version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"1.13.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Go version \(using buildpack.yml\): 1\.13\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Go 1\.13\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			secondContainer, err = docker.Container.Run.WithCommand("go run main.go").Execute(secondImage.ID)
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
