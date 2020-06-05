package gocompiler_test

import (
	"testing"

	gocompiler "github.com/paketo-buildpacks/go-compiler"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		resolver gocompiler.PlanEntryResolver
	)

	it.Before(func() {
		resolver = gocompiler.NewPlanEntryResolver()
	})

	context("Resolve", func() {
		it("chooses a plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"number": 1,
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"number": 2,
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"number": 1,
				},
			}))
		})

		it("merges the build and launch flags", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"build":  true,
						"number": 1,
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"launch": true,
						"number": 2,
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
					"number": 1,
				},
			}))
		})

		it("prioritizes versions from buildpack.yml", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"number": 1,
					},
				},
				{
					Name: "go",
					Metadata: map[string]interface{}{
						"number":         2,
						"version-source": "buildpack.yml",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"number":         2,
					"version-source": "buildpack.yml",
				},
			}))
		})
	})
}
