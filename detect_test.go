package godist_test

import (
	"os"
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
	}, spec.Sequential())
}
