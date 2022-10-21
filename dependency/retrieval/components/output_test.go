package components_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOutput(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect    = NewWithT(t).Expect
		outputDir string
		depDate   time.Time
	)

	it.Before(func() {
		var err error
		outputDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		depDate, err = time.ParseInLocation("2006-01-02", "2024-11-12", time.UTC)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(outputDir)).To(Succeed())
	})

	context("WriteOutput", func() {
		it("will write an output file", func() {
			err := components.WriteOutput(filepath.Join(outputDir, "output.json"), []cargo.ConfigMetadataDependency{
				{
					DeprecationDate: &depDate,
					Licenses:        []interface{}{"MIT", "MIT-0"},
					Name:            ".NET Core SDK",
					SHA256:          "",
				},
			}, "target")
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(outputDir, "output.json")).To(BeAFileMatching("[{\"deprecation_date\":\"2024-11-12T00:00:00Z\",\"licenses\":[\"MIT\",\"MIT-0\"],\"name\":\".NET Core SDK\",\"target\":\"target\"}]\n"))
		})

		context("failure cases", func() {
			context("the output file cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(outputDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(outputDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := components.WriteOutput(filepath.Join(outputDir, "output.json"), []cargo.ConfigMetadataDependency{
						{
							DeprecationDate: &depDate,
							Licenses:        []interface{}{"MIT", "MIT-0"},
							Name:            ".NET Core SDK",
							SHA256:          "",
						},
					}, "target")
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
