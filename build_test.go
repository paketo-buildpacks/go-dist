package godist_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	godist "github.com/paketo-buildpacks/go-dist"
	"github.com/paketo-buildpacks/go-dist/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		cnbDir            string
		timestamp         time.Time
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "go",
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "go",
			Name:    "go-dependency-name",
			SHA256:  "go-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "go-dependency-uri",
			Version: "go-dependency-version",
		}

		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "go",
				Metadata: packit.BOMMetadata{
					Version: "go-dependency-version",
					Checksum: packit.BOMChecksum{
						Algorithm: packit.SHA256,
						Hash:      "go-dependency-sha",
					},
					URI: "go-dependency-uri",
				},
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := godist.NewGoLogger(scribe.NewEmitter(buffer))

		build = godist.Build(entryResolver, dependencyManager, clock, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs go", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "go"},
				},
			},
			Platform: packit.Platform{Path: "platform"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name:             "go",
					Path:             filepath.Join(layersDir, "go"),
					SharedEnv:        packit.Environment{},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					Metadata: map[string]interface{}{
						"dependency-sha": "go-dependency-sha",
						"built_at":       timestamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.Name).To(Equal("go"))
		Expect(entryResolver.ResolveCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "go"},
		}))

		Expect(entryResolver.MergeLayerTypesCall.Receives.Name).To(Equal("go"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "go"},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("go"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("default"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "go",
			Name:    "go-dependency-name",
			SHA256:  "go-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "go-dependency-uri",
			Version: "go-dependency-version",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "go")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:      "go",
				Name:    "go-dependency-name",
				SHA256:  "go-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "go-dependency-uri",
				Version: "go-dependency-version",
			},
		}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Go version"))
		Expect(buffer.String()).To(ContainSubstring("Selected go-dependency-name version (using <unknown>): go-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing Go go-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Completed in"))
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it("makes the layer available in those phases", func() {
			result, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "go",
							Metadata: map[string]interface{}{
								"build":  true,
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name:             "go",
						Path:             filepath.Join(layersDir, "go"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						Metadata: map[string]interface{}{
							"dependency-sha": "go-dependency-sha",
							"built_at":       timestamp.Format(time.RFC3339Nano),
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "go",
							Metadata: packit.BOMMetadata{
								Version: "go-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "go-dependency-sha",
								},
								URI: "go-dependency-uri",
							},
						},
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "go",
							Metadata: packit.BOMMetadata{
								Version: "go-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "go-dependency-sha",
								},
								URI: "go-dependency-uri",
							},
						},
					},
				},
			}))
		})
	})

	context("when the plan entry selected has buildpack.yml as its source", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
				},
			}
		})

		it("prints a warning message", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "go",
							Metadata: map[string]interface{}{
								"build":  true,
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name:             "go",
						Path:             filepath.Join(layersDir, "go"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            false,
						Launch:           false,
						Cache:            false,
						Metadata: map[string]interface{}{
							"dependency-sha": "go-dependency-sha",
							"built_at":       timestamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Go version"))
			Expect(buffer.String()).To(ContainSubstring("Selected go-dependency-name version (using buildpack.yml): go-dependency-version"))
			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the Go Dist version through buildpack.yml will be deprecated soon in Go Dist Buildpack v2.0.0."))
			Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_GO_VERSION environment variable instead. See README.md for more information."))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Installing Go go-dependency-version"))
			Expect(buffer.String()).To(ContainSubstring("Completed in"))
		})
	})

	context("failure cases", func() {
		context("when the dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "go"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when the go layer cannot be retrieved", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layersDir, "go.toml"), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "go"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when the go layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "go", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "go"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "go"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "go"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})
		})

		context("when the dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("failed to deliver dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "go"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError("failed to deliver dependency"))
			})
		})
	})
}
