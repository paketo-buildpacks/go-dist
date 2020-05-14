package main

import (
	"sort"

	"github.com/paketo-buildpacks/packit"
)

type PlanEntryResolver struct{}

func NewPlanEntryResolver() PlanEntryResolver {
	return PlanEntryResolver{}
}

func (r PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	priorities := map[string]int{
		"buildpack.yml": 1,
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

	return entry
}
