package godist_test

import (
	"os"
	"path/filepath"
	"testing"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect packit.DetectFunc
	)

	it.Before(func() {
		detect = godist.Detect()
	})

	it("returns a plan that provides go", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "go"},
				},
			},
		}))

	})

	context("when the BP_GO_VERSION is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_GO_VERSION", "some-version")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_GO_VERSION")).To(Succeed())
		})

		it("returns a plan that requires that version of go", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "go"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "go",
							Metadata: godist.BuildPlanMetadata{
								VersionSource: "BP_GO_VERSION",
								Version:       "some-version",
							},
						},
					},
				},
			}))

		})
	})

	context("when there is a buildpack.yml", func() {
		var workingDir string
		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, os.ModePerm))
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("fails the build with a deprecation notice", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError("working directory contains deprecated 'buildpack.yml'; use environment variables for configuration"))
		})

		context("and its contents cannot be read", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})
			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
