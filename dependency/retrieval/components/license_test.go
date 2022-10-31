package components_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLicense(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("GenerateLicenseInformation", func() {
		var (
			server *httptest.Server
		)

		it.Before(func() {
			buffer := bytes.NewBuffer(nil)
			gw := gzip.NewWriter(buffer)
			tw := tar.NewWriter(gw)

			Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err := tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			licenseFile := "some-dir/LICENSE.txt"
			Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(lFile))})).To(Succeed())
			_, err = tw.Write([]byte(lFile))
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())
			Expect(gw.Close()).To(Succeed())

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "/":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(buffer.Bytes())
					Expect(err).NotTo(HaveOccurred())

				case "/bad-archive":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
					Expect(err).NotTo(HaveOccurred())

				default:
					t.Fatalf("unknown path: %s", req.URL.Path)
				}
			}))

		})

		it("returns returns a cargo dependency generated from the given release", func() {
			licenses, err := components.GenerateLicenseInformation(server.URL)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(Equal([]interface{}{"MIT", "MIT-0"}))
		})

		context("failure cases", func() {
			context("when the url is invalid", func() {
				it("returns an error", func() {
					_, err := components.GenerateLicenseInformation("invalid url")
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("when the artifact is not a supported archive type", func() {
				it("returns an error", func() {
					_, err := components.GenerateLicenseInformation(fmt.Sprintf("%s/bad-archive", server.URL))
					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})
		})
	})
}
