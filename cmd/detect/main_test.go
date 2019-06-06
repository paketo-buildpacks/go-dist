package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/detect"
	"github.com/cloudfoundry/go-cnb/golang"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	it("always passes", func() {
		code, err := runDetect(factory.Detect)
		Expect(err).NotTo(HaveOccurred())

		Expect(code).To(Equal(detect.PassStatusCode))

		Expect(factory.Output).To(Equal(buildplan.BuildPlan{
			golang.Dependency: buildplan.Dependency{
				Version:  "",
				Metadata: buildplan.Metadata{"build": true, "launch": false},
			}}))
	})

	when("testing versions", func() {

		when("there is no buildpack.yml", func() {
			it("shouldn't set the version in the buildplan", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					golang.Dependency: buildplan.Dependency{
						Version:  "",
						Metadata: buildplan.Metadata{"build": true, "launch": false},
					},
				}))
			})
		})

		when("there is a buildpack.yml", func() {
			const version string = "1.2.3"

			it.Before(func() {
				buildpackYAMLString := fmt.Sprintf("go:\n  version: %s", version)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of golang", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					golang.Dependency: buildplan.Dependency{
						Version:  version,
						Metadata: buildplan.Metadata{"build": true, "launch": false},
					},
				}))
			})
		})

		when("there is a is an existing version from the build plan and a buildpack.yml", func() {
			const buildpackYAMLVersion string = "1.2.3"
			const existingVersion string = "4.5.6"

			it.Before(func() {
				factory.AddBuildPlan(golang.Dependency, buildplan.Dependency{
					Version: existingVersion,
				})

				buildpackYAMLString := fmt.Sprintf("go:\n  version: %s", buildpackYAMLVersion)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of golang defined in buildpack.yml", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				Expect(factory.Output).To(Equal(buildplan.BuildPlan{
					golang.Dependency: buildplan.Dependency{
						Version:  buildpackYAMLVersion,
						Metadata: buildplan.Metadata{"build": true, "launch": false},
					},
				}))
			})
		})
	})
}
