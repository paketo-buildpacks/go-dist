package integration_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/cloudfoundry/occam"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	buildpack        string
	offlineBuildpack string
	noAppBuildpack   string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())

	buildpack, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	offlineBuildpack, _, err = dagger.PackageCachedBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	noAppBuildpack, err = dagger.GetLatestCommunityBuildpack("ForestEckhardt", "no-app-cnb")
	Expect(err).NotTo(HaveOccurred())

	// HACK: we need to fix dagger and the package.sh scripts so that this isn't required
	buildpack = fmt.Sprintf("%s.tgz", buildpack)
	offlineBuildpack = fmt.Sprintf("%s.tgz", offlineBuildpack)

	defer func() {
		dagger.DeleteBuildpack(buildpack)
		dagger.DeleteBuildpack(offlineBuildpack)
		dagger.DeleteBuildpack(noAppBuildpack)
	}()

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}))
	suite("BuildpackYAML", testBuildpackYAML, spec.Parallel())
	suite("Default", testDefault, spec.Parallel())
	suite("LayerReuse", testLayerReuse)
	suite("Offline", testOffline)
	suite.Run(t)
}

func ContainerLogs(id string) func() string {
	docker := occam.NewDocker()

	return func() string {
		logs, _ := docker.Container.Logs.Execute(id)
		return logs.String()
	}
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})
	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	err = gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--tags", strings.TrimSpace(revListOut.String())},
		Stdout: stdout,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(stdout.String(), "v")), nil
}
