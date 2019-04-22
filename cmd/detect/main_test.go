package main

import (
	"testing"

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
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("always passes", func() {
		f := test.NewDetectFactory(t)
		code, err := runDetect(f.Detect)
		Expect(err).NotTo(HaveOccurred())

		Expect(code).To(Equal(detect.PassStatusCode))

		Expect(f.Output).To(Equal(buildplan.BuildPlan{
			golang.Dependency: buildplan.Dependency{
				Version:  "",
				Metadata: buildplan.Metadata{"build": true, "launch": false},
			}}))
	})

}
