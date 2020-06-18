package godist

import (
	"io"
	"strconv"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

type LogEmitter struct {
	scribe.Logger
}

func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Logger: scribe.NewLogger(output),
	}
}

func (l LogEmitter) Title(info packit.BuildpackInfo) {
	l.Logger.Title("%s %s", info.Name, info.Version)
}

func (l LogEmitter) Candidates(entries []packit.BuildpackPlanEntry) {
	l.Subprocess("Candidate version sources (in priority order):")

	var (
		sources [][2]string
		maxLen  int
	)

	for _, entry := range entries {
		versionSource, ok := entry.Metadata["version-source"].(string)
		if !ok {
			versionSource = "<unknown>"
		}

		if len(versionSource) > maxLen {
			maxLen = len(versionSource)
		}

		sources = append(sources, [2]string{versionSource, entry.Version})
	}

	for _, source := range sources {
		l.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	l.Break()
}

func (l LogEmitter) SelectedDependency(entry packit.BuildpackPlanEntry, dependency postal.Dependency, now time.Time) {
	versionSource, ok := entry.Metadata["version-source"]
	if !ok {
		versionSource = "<unknown>"
	}

	l.Subprocess("Selected %s version (using %s): %s", dependency.Name, versionSource, dependency.Version)

	if (dependency.DeprecationDate != time.Time{}) {
		deprecationDate := dependency.DeprecationDate
		switch {
		case (deprecationDate.Add(-30*24*time.Hour).Before(now) && deprecationDate.After(now)):
			l.Action("Version %s of %s will be deprecated after %s.", dependency.Version, dependency.Name, dependency.DeprecationDate.Format("2006-01-02"))
			l.Action("Migrate your application to a supported version of %s before this time.", dependency.Name)
		case (deprecationDate == now || deprecationDate.Before(now)):
			l.Action("Version %s of %s is deprecated.", dependency.Version, dependency.Name)
			l.Action("Migrate your application to a supported version of %s.", dependency.Name)
		}
	}

	l.Break()
}
