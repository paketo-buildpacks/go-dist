package godist

import (
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type GoLogger struct {
	scribe.Emitter
}

func NewGoLogger(emitter scribe.Emitter) GoLogger {
	return GoLogger{Emitter: emitter}
}
