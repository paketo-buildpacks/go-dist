package godist

const (
	DependencySHAKey = "dependency-sha"
	GoDependency     = "go"
	GoLayerName      = "go"
)

var Priorities = []interface{}{
	"BP_GO_VERSION",
	"buildpack.yml",
	"go.mod",
}
