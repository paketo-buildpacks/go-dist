package godist

import (
	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type GoLogger struct {
	scribe.Emitter
}

func NewGoLogger(emitter scribe.Emitter) GoLogger {
	return GoLogger{Emitter: emitter}
}

func (gl GoLogger) WarnBuildpackYML(buildpackVersion string) {
	nextMajorVersion := semver.MustParse(buildpackVersion).IncMajor()
	gl.Subprocess("WARNING: Setting the Go Dist version through buildpack.yml will be deprecated soon in Go Dist Buildpack v%s.", nextMajorVersion.String())
	gl.Subprocess("Please specify the version through the $BP_GO_VERSION environment variable instead. See README.md for more information.")
	gl.Break()
}
