# Go Distribution Cloud Native Buildpack

The Go Distribution CNB provides the Go binary distribution that can be used to
execute [Go tooling](https://golang.org/cmd/go/). The buildpack installs the Go
binary distribution onto the `$PATH` which makes it available for subsequent
buildpacks. These buildpacks can then use that distribution to run Go tooling
including building Go application binaries. Examples of buildpacks that perform
this binary building process include the [Go Mod
CNB](https://github.com/paketo-buildpacks/go-mod) and the [Dep
CNB](https://github.com/paketo-buildpacks/dep).

## Integration

The Go Distribution CNB provides Go as a dependency. Downstream buildpacks, like
[Go Mod](https://github.com/paketo-buildpacks/go-mod) or
[Dep](https://github.com/paketo-buildpacks/dep), can require the go dependency
by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the  dependency is "go". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "go"

  # The version of the Go dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "1.*", "1.13.*", or even
  # "1.13.9".
  version = "1.13.9"

  # The Go buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Go
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Go
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the Go
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run Go at runtime, this flag should
    # be set to true.
    launch = true
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.

## Go Build Configuration

To configure the Go version, please use the `BP_GO_VERSION` environment
variable at build time either directly
(ex. `pack build my-app --env BP_GO_VERSION=~1.14.1`) or through a
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
