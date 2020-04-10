package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/paketo-buildpacks/go-compiler/golang"

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

		provided := []buildplan.Provided{{Name: golang.Dependency}}
		required := []buildplan.Required{{
			Name:     golang.Dependency,
			Version:  "",
			Metadata: buildplan.Metadata{"build": true, "launch": false},
		}}

		Expect(factory.Plans.Plan.Provides).To(Equal(provided))
		Expect(factory.Plans.Plan.Requires).To(Equal(required))
	})

	when("testing versions", func() {

		when("there is no buildpack.yml", func() {
			it("shouldn't set the version in the buildplan", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				provided := []buildplan.Provided{{Name: golang.Dependency}}
				required := []buildplan.Required{{
					Name:     golang.Dependency,
					Version:  "",
					Metadata: buildplan.Metadata{"build": true, "launch": false},
				}}

				Expect(factory.Plans.Plan.Provides).To(Equal(provided))
				Expect(factory.Plans.Plan.Requires).To(Equal(required))

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

				provided := []buildplan.Provided{{Name: golang.Dependency}}
				required := []buildplan.Required{{
					Name:     golang.Dependency,
					Version:  version,
					Metadata: buildplan.Metadata{"build": true, "launch": false},
				}}

				Expect(factory.Plans.Plan.Provides).To(Equal(provided))
				Expect(factory.Plans.Plan.Requires).To(Equal(required))
			})
		})

		when("there is a buildpack.yml with additional values", func() {
			const version string = "1.2.3"

			it.Before(func() {
				buildpackYAMLString := fmt.Sprintf("go:\n  version: %s\nfoo: bar", version)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of golang", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				provided := []buildplan.Provided{{Name: golang.Dependency}}
				required := []buildplan.Required{{
					Name:     golang.Dependency,
					Version:  version,
					Metadata: buildplan.Metadata{"build": true, "launch": false},
				}}

				Expect(factory.Plans.Plan.Provides).To(Equal(provided))
				Expect(factory.Plans.Plan.Requires).To(Equal(required))
			})
		})

		when("there is a buildpack.yml with value related to another buildpack", func() {

			it.Before(func() {
				buildpackYAMLString := "foo: bar"
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of golang", func() {
				code, err := runDetect(factory.Detect)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))

				provided := []buildplan.Provided{{Name: golang.Dependency}}
				required := []buildplan.Required{{
					Name:     golang.Dependency,
					Version:  "",
					Metadata: buildplan.Metadata{"build": true, "launch": false},
				}}

				Expect(factory.Plans.Plan.Provides).To(Equal(provided))
				Expect(factory.Plans.Plan.Requires).To(Equal(required))

			})
		})
	})
}
