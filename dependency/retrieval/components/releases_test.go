package components_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/go-dist/dependency/retrieval/components"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReleases(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("Fetcher", func() {
		var (
			fetcher components.Fetcher

			server *httptest.Server
		)

		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "/":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, `[
   {
  "version": "go1.19",
  "stable": true,
  "files": [
   {
    "filename": "go1.19.src.tar.gz",
    "os": "",
    "arch": "",
    "version": "go1.19",
    "sha256": "9419cc70dc5a2523f29a77053cafff658ed21ef3561d9b6b020280ebceab28b9",
    "size": 26521849,
    "kind": "source"
   }
  ]
 },
 {
  "version": "go1.19rc2",
  "stable": false,
  "files": [
   {
    "filename": "go1.19rc2.src.tar.gz",
    "os": "",
    "arch": "",
    "version": "go1.19rc2",
    "sha256": "c68d7019a6a0b9852ae0e96d7e7deb772492a23272fd6d13afe05b40c912e51b",
    "size": 26593323,
    "kind": "source"
   }
  ]
 }
]
`)

				case "/non-200":
					w.WriteHeader(http.StatusTeapot)

				case "/no-parse":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `???`)

				case "/no-version-parse":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, `[
   {
  "version": "invalid semver",
  "stable": true
 }]
`)

				default:
					t.Fatalf("unknown path: %s", req.URL.Path)
				}
			}))

			fetcher = components.NewFetcher().WithReleasePage(server.URL)
		})

		it("fetches a list of relevant releases", func() {
			releases, err := fetcher.Get()
			Expect(err).NotTo(HaveOccurred())

			Expect(releases).To(Equal([]components.Release{
				{
					SemVer:  semver.MustParse("1.19"),
					Version: "go1.19",
					Stable:  true,
					Files: []components.ReleaseFile{
						{
							URL:      "https://go.dev/dl/go1.19.src.tar.gz",
							Filename: "go1.19.src.tar.gz",
							OS:       "",
							Arch:     "",
							Version:  "go1.19",
							SHA256:   "9419cc70dc5a2523f29a77053cafff658ed21ef3561d9b6b020280ebceab28b9",
							Size:     26521849,
							Kind:     "source",
						},
					},
				},
			}))
		})

		context("failure cases", func() {
			context("when the release page get fails", func() {
				it.Before(func() {
					fetcher = fetcher.WithReleasePage("not a valid URL")
				})

				it("returns an error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("when the release page returns non 200 code", func() {
				it.Before(func() {
					fetcher = fetcher.WithReleasePage(fmt.Sprintf("%s/non-200", server.URL))
				})

				it("returns an error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(fmt.Sprintf("received a non 200 status code from %s: status code 418 received", fmt.Sprintf("%s/non-200", server.URL))))
				})
			})

			context("when the release page cannot be parsed", func() {
				it.Before(func() {
					fetcher = fetcher.WithReleasePage(fmt.Sprintf("%s/no-parse", server.URL))
				})

				it("returns an error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("invalid character '?' looking for beginning of value")))
				})
			})

			context("when the release page has unparsable version", func() {
				it.Before(func() {
					fetcher = fetcher.WithReleasePage(fmt.Sprintf("%s/no-version-parse", server.URL))
				})

				it("returns an error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("invalid semantic version")))
				})
			})
		})
	})
}
