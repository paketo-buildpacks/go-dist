package main_test

import (
	"testing"

	main "github.com/paketo-buildpacks/go-compiler"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildPlanRefinery(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		refinery main.BuildPlanRefinery
	)

	it.Before(func() {
		refinery = main.NewBuildPlanRefinery()
	})

	context("BillOfMaterials", func() {
		it("returns a refined build plan entry", func() {
			entry := refinery.BillOfMaterials(postal.Dependency{
				ID:      "some-id",
				Name:    "some-name",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri",
				SHA256:  "some-sha",
				Version: "some-version",
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "some-id",
				Version: "some-version",
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     "some-name",
					"sha256":   "some-sha",
					"stacks":   []string{"some-stack"},
					"uri":      "some-uri",
				},
			}))
		})
	})
}
