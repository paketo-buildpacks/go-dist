package godist_test

import (
	"bytes"
	"testing"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer   *bytes.Buffer
		resolver godist.PlanEntryResolver
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		resolver = godist.NewPlanEntryResolver(godist.NewLogEmitter(buffer))
	})

	context("when a BP_GO_VERSION entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "go-mod-version",
						"version-source": "go.mod",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "BP_GO_VERSION-version",
						"version-source": "BP_GO_VERSION",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "buildpack-yml-version",
						"version-source": "buildpack.yml",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"version":        "BP_GO_VERSION-version",
					"version-source": "BP_GO_VERSION",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      BP_GO_VERSION -> \"BP_GO_VERSION-version\""))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      go.mod        -> \"go-mod-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
		})
	})

	context("when a buildpack.yml entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "go-mod-version",
						"version-source": "go.mod",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "buildpack-yml-version",
						"version-source": "buildpack.yml",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"version":        "buildpack-yml-version",
					"version-source": "buildpack.yml",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      go.mod        -> \"go-mod-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
		})
	})

	context("when a buildpack.yml entry is not included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "go-mod-version",
						"version-source": "go.mod",
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"version":        "go-mod-version",
					"version-source": "go.mod",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      go.mod    -> \"go-mod-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown> -> \"other-version\""))
		})
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name: "go",
						Metadata: map[string]interface{}{
							"version":        "go-mod-version",
							"version-source": "go.mod",
						},
					},
					{
						Name: "go",
						Metadata: map[string]interface{}{
							"version":        "other-version",
							"version-source": "other-source",
							"build":          true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "go",
					Metadata: map[string]interface{}{
						"version":        "go-mod-version",
						"version-source": "go.mod",
						"build":          true,
					},
				}))
			})
		})
	})

	context("when an unknown source entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"version": "other-version",
				},
			}))
		})
	})
}
