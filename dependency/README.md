# Dependency

Pre-compiled distributions of golang are provided by google and are supported on any stacks.

This directory contains the script to facilitate the following:
* Identifying when there is a new version of golang available
* Create json file used as input by jam to update dependency versions in buidpack.toml

## Running locally

Running the steps locally can be useful for iterating on the compilation process
(e.g. changing compilation options) as well as debugging.

### Retrieval

Retrieve latest versions with the following. This takes the current buildpack.toml as input and will only add new versions to the json file it produces. For testing purposes you can remove all existing versions from buildpack.toml while leaving the constraint definitions and it will generate all new versions.

```
cd ./retrieval

go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output ../../metadata.json
```

See [retrieval/README.md](retrieval/README.md) for more details.

### Updating buildpack.toml

Update bulidpack.toml from metadata.json with the following command:

* This assumes you are running from the root of the repo
  * `cd ../..`
* jam is required. You can run `scripts/package.sh -v SOME_SEMVER` and it will install jam to `.bin/jam`
  * Change `SOME_SEMVER` to something like `0.0.0`

```
.bin/jam update-dependencies \
  --buildpack-file buildpack.toml \
  --metadata-file metadata.json
```

It will produce the output like this:

```
Updating buildpack.toml with new versions:  [1.24.6 1.24.7 1.25.0 1.25.1]
```