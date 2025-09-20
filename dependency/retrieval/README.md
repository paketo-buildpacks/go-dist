# Dependency Retrieval

## Running locally

Run the following command:

```
go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output ../../metadata.json
```

Example output (abbreviated for clarity):

```
Found 251 versions of go from upstream
[
  "1.25.1",  ... "1.0.0"
]
Found 2 versions of go for constraint 1.25.*
[
  "1.25.1", "1.25.0"
]
Found 8 versions of go for constraint 1.24.*
[
  "1.24.7", "1.24.6", "1.24.5", "1.24.4", "1.24.3",
  "1.24.2", "1.24.1", "1.24.0"
]
Found 2 versions of go newer than '' for constraint 1.25.*, after limiting for 2 patches
[
  "1.25.1", "1.25.0"
]
Found 2 versions of go newer than '' for constraint 1.24.*, after limiting for 2 patches
[
  "1.24.7", "1.24.6"
]
Found 4 versions of go as new versions
[
  "1.25.1", "1.25.0", "1.24.7", "1.24.6"
]
Generating metadata for 1.25.1, platform linux/amd64, with stacks [*]
Generating metadata for 1.25.0, platform linux/amd64, with stacks [*]
Generating metadata for 1.24.7, platform linux/amd64, with stacks [*]
Generating metadata for 1.24.6, platform linux/amd64, with stacks [*]
Generating metadata for 1.25.1, platform linux/arm64, with stacks [*]
Generating metadata for 1.25.0, platform linux/arm64, with stacks [*]
Generating metadata for 1.24.7, platform linux/arm64, with stacks [*]
Generating metadata for 1.24.6, platform linux/arm64, with stacks [*]
Wrote metadata to ../../metadata.json

```