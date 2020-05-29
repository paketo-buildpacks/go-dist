package main_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	main "github.com/paketo-buildpacks/go-compiler"
	"github.com/paketo-buildpacks/go-compiler/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
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
		planRefinery      *fakes.PlanRefinery
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
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

		planRefinery = &fakes.PlanRefinery{}
		planRefinery.BillOfMaterialsCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name:    "go",
			Version: "go-dependency-version",
			Metadata: map[string]interface{}{
				"name":   "go-dependency-name",
				"sha256": "go-dependency-sha",
				"stacks": []string{"some-stack"},
				"uri":    "go-dependency-uri",
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := main.NewLogEmitter(buffer)

		build = main.Build(entryResolver, dependencyManager, planRefinery, clock, logEmitter)
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
			Layers: packit.Layers{Path: layersDir},
			Stack:  "some-stack",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name:    "go",
						Version: "go-dependency-version",
						Metadata: map[string]interface{}{
							"name":   "go-dependency-name",
							"sha256": "go-dependency-sha",
							"stacks": []string{"some-stack"},
							"uri":    "go-dependency-uri",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name:      "go",
					Path:      filepath.Join(layersDir, "go"),
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    false,
					Cache:     false,
					Metadata: map[string]interface{}{
						"dependency-sha": "go-dependency-sha",
						"built_at":       timestamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "go"},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("go"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "go",
			Name:    "go-dependency-name",
			SHA256:  "go-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "go-dependency-uri",
			Version: "go-dependency-version",
		}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "go")))

		Expect(planRefinery.BillOfMaterialsCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "go",
			Name:    "go-dependency-name",
			SHA256:  "go-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "go-dependency-uri",
			Version: "go-dependency-version",
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
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "go",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
				},
			}
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
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name:    "go",
							Version: "go-dependency-version",
							Metadata: map[string]interface{}{
								"name":   "go-dependency-name",
								"sha256": "go-dependency-sha",
								"stacks": []string{"some-stack"},
								"uri":    "go-dependency-uri",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "go",
						Path:      filepath.Join(layersDir, "go"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     true,
						Launch:    true,
						Cache:     true,
						Metadata: map[string]interface{}{
							"dependency-sha": "go-dependency-sha",
							"built_at":       timestamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
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
				err := ioutil.WriteFile(filepath.Join(layersDir, "go.toml"), nil, 0000)
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
				Expect(os.Chmod(filepath.Join(layersDir, "go"), 0000)).To(Succeed())
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
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
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
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})
	})
}
