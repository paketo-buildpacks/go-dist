package fakes

import (
	"sync"

	packit "github.com/paketo-buildpacks/packit/v2"
)

type EntryResolver struct {
	MergeLayerTypesCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Name    string
			Entries []packit.BuildpackPlanEntry
		}
		Returns struct {
			Launch bool
			Build  bool
		}
		Stub func(string, []packit.BuildpackPlanEntry) (bool, bool)
	}
	ResolveCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Name      string
			Entries   []packit.BuildpackPlanEntry
			Priorites []interface {
			}
		}
		Returns struct {
			BuildpackPlanEntry      packit.BuildpackPlanEntry
			BuildpackPlanEntrySlice []packit.BuildpackPlanEntry
		}
		Stub func(string, []packit.BuildpackPlanEntry, []interface {
		}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	}
}

func (f *EntryResolver) MergeLayerTypes(param1 string, param2 []packit.BuildpackPlanEntry) (bool, bool) {
	f.MergeLayerTypesCall.mutex.Lock()
	defer f.MergeLayerTypesCall.mutex.Unlock()
	f.MergeLayerTypesCall.CallCount++
	f.MergeLayerTypesCall.Receives.Name = param1
	f.MergeLayerTypesCall.Receives.Entries = param2
	if f.MergeLayerTypesCall.Stub != nil {
		return f.MergeLayerTypesCall.Stub(param1, param2)
	}
	return f.MergeLayerTypesCall.Returns.Launch, f.MergeLayerTypesCall.Returns.Build
}
func (f *EntryResolver) Resolve(param1 string, param2 []packit.BuildpackPlanEntry, param3 []interface {
}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry) {
	f.ResolveCall.mutex.Lock()
	defer f.ResolveCall.mutex.Unlock()
	f.ResolveCall.CallCount++
	f.ResolveCall.Receives.Name = param1
	f.ResolveCall.Receives.Entries = param2
	f.ResolveCall.Receives.Priorites = param3
	if f.ResolveCall.Stub != nil {
		return f.ResolveCall.Stub(param1, param2, param3)
	}
	return f.ResolveCall.Returns.BuildpackPlanEntry, f.ResolveCall.Returns.BuildpackPlanEntrySlice
}
