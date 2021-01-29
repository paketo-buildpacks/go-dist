package godist

import (
	"sort"

	"github.com/paketo-buildpacks/packit"
)

type PlanEntryResolver struct {
	logger LogEmitter
}

func NewPlanEntryResolver(logger LogEmitter) PlanEntryResolver {
	return PlanEntryResolver{
		logger: logger,
	}
}

func (r PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	priorities := map[string]int{
		"BP_GO_VERSION": 3,
		"buildpack.yml": 2,
		"go.mod":        1,
		"":              -1,
	}

	sort.Slice(entries, func(i, j int) bool {
		left, _ := entries[i].Metadata["version-source"].(string)
		right, _ := entries[j].Metadata["version-source"].(string)

		return priorities[left] > priorities[right]
	})

	entry := entries[0]
	if entry.Metadata == nil {
		entry.Metadata = map[string]interface{}{}
	}

	for _, e := range entries {
		for _, phase := range []string{"build", "launch"} {
			if e.Metadata[phase] == true {
				entry.Metadata[phase] = true
			}
		}
	}

	r.logger.Candidates(entries)

	return entry
}
