package components_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

const (
	lFile = `The MIT License (MIT)

Copyright (c) .NET Foundation and Contributors

All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
)

func testDependency(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("ConvertReleaseToDependeny", func() {
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
				case "/source":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(buffer.Bytes())
					Expect(err).NotTo(HaveOccurred())

				case "/archive":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(nil)
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
			dependency, err := components.ConvertReleaseToDependency(components.Release{
				SemVer:  semver.MustParse("1.19"),
				Version: "go1.19",
				Files: []components.ReleaseFile{
					{
						URL:      fmt.Sprintf("%s/source", server.URL),
						Filename: "go1.19.src.tar.gz",
						OS:       "",
						Arch:     "",
						Version:  "go1.19",
						SHA256:   "ad1b820bde8c32707f8bb8ce636750b1c1b7c83a82e43481910bef2f4f77dcb5",
						Size:     26521849,
						Kind:     "source",
					},
					{
						URL:      fmt.Sprintf("%s/archive", server.URL),
						Filename: "go1.19.linux-amd64.tar.gz",
						OS:       "linux",
						Arch:     "amd64",
						Version:  "go1.19",
						SHA256:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
						Size:     26521849,
						Kind:     "archive",
					},
				},
			},
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				Checksum:        "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				CPE:             "cpe:2.3:a:golang:go:1.19:*:*:*:*:*:*:*",
				PURL:            fmt.Sprintf("pkg:generic/go@go1.19?checksum=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855&download_url=%s", fmt.Sprintf("%s/archive", server.URL)),
				ID:              "go",
				Licenses:        []interface{}{"MIT", "MIT-0"},
				Name:            "Go",
				SHA256:          "",
				Source:          fmt.Sprintf("%s/source", server.URL),
				SourceChecksum:  "sha256:ad1b820bde8c32707f8bb8ce636750b1c1b7c83a82e43481910bef2f4f77dcb5",
				SourceSHA256:    "",
				Stacks:          []string{"*"},
				StripComponents: 1,
				URI:             fmt.Sprintf("%s/archive", server.URL),
				Version:         "1.19.0",
			}))
		})

		context("failure cases", func() {
			context("when there is not a release files", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{})
					Expect(err).To(MatchError("could not find release file for linux/amd64"))
				})
			})

			context("when the source is not a supported archive type", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("1.19"),
						Version: "go1.19",
						Files: []components.ReleaseFile{
							{
								URL:      fmt.Sprintf("%s/bad-archive", server.URL),
								Filename: "go1.19.src.tar.gz",
								OS:       "",
								Arch:     "",
								Version:  "go1.19",
								SHA256:   "5a95bcffa592dcc7689ef5b4d993da3ca805b3c58d1710da8effeedbda87d471",
								Size:     26521849,
								Kind:     "source",
							},
							{
								URL:      fmt.Sprintf("%s/archive", server.URL),
								Filename: "go1.19.linux-amd64.tar.gz",
								OS:       "linux",
								Arch:     "amd64",
								Version:  "go1.19",
								SHA256:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
								Size:     26521849,
								Kind:     "archive",
							},
						},
					},
					)
					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})

			context("when the checksum does not match", func() {
				context("when the artifact file has the wrong checksum", func() {
					it("returns an error", func() {
						_, err := components.ConvertReleaseToDependency(components.Release{
							SemVer:  semver.MustParse("1.19"),
							Version: "go1.19",
							Files: []components.ReleaseFile{
								{
									URL:      fmt.Sprintf("%s/source", server.URL),
									Filename: "go1.19.src.tar.gz",
									OS:       "",
									Arch:     "",
									Version:  "go1.19",
									SHA256:   "ad1b820bde8c32707f8bb8ce636750b1c1b7c83a82e43481910bef2f4f77dcb5",
									Size:     26521849,
									Kind:     "source",
								},
								{
									URL:      fmt.Sprintf("%s/archive", server.URL),
									Filename: "go1.19.linux-amd64.tar.gz",
									OS:       "linux",
									Arch:     "amd64",
									Version:  "go1.19",
									SHA256:   "invalid checksum",
									Size:     26521849,
									Kind:     "archive",
								},
							},
						},
						)
						Expect(err).To(MatchError("the given checksum of the artifact does not match with downloaded artifact"))
					})
				})

				context("when the source file has the wrong checksum", func() {
					it("returns an error", func() {
						_, err := components.ConvertReleaseToDependency(components.Release{
							SemVer:  semver.MustParse("1.19"),
							Version: "go1.19",
							Files: []components.ReleaseFile{
								{
									URL:      fmt.Sprintf("%s/source", server.URL),
									Filename: "go1.19.src.tar.gz",
									OS:       "",
									Arch:     "",
									Version:  "go1.19",
									SHA256:   "invalid checksum",
									Size:     26521849,
									Kind:     "source",
								},
								{
									URL:      fmt.Sprintf("%s/archive", server.URL),
									Filename: "go1.19.linux-amd64.tar.gz",
									OS:       "linux",
									Arch:     "amd64",
									Version:  "go1.19",
									SHA256:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
									Size:     26521849,
									Kind:     "archive",
								},
							},
						},
						)
						Expect(err).To(MatchError("the given checksum of the source does not match with downloaded source"))
					})
				})
			})
		})
	})
}
