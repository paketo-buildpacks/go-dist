package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type PlanRefinery struct {
	BillOfMaterialsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency postal.Dependency
		}
		Returns struct {
			BuildpackPlanEntry packit.BuildpackPlanEntry
		}
		Stub func(postal.Dependency) packit.BuildpackPlanEntry
	}
}

func (f *PlanRefinery) BillOfMaterials(param1 postal.Dependency) packit.BuildpackPlanEntry {
	f.BillOfMaterialsCall.Lock()
	defer f.BillOfMaterialsCall.Unlock()
	f.BillOfMaterialsCall.CallCount++
	f.BillOfMaterialsCall.Receives.Dependency = param1
	if f.BillOfMaterialsCall.Stub != nil {
		return f.BillOfMaterialsCall.Stub(param1)
	}
	return f.BillOfMaterialsCall.Returns.BuildpackPlanEntry
}
