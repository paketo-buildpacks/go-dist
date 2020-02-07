#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# shellcheck source=.util/tools.sh
source "${PWD}/scripts/.util/tools.sh"

util::tools::install \
    --directory "${PWD}/.bin" \
    --pack-version "latest"

PACKAGE_DIR=${PACKAGE_DIR:-"${PWD##*/}_$(openssl rand -hex 4)"}

full_path=$(realpath "$PACKAGE_DIR")
args=".bin/packager -uncached"

while getopts "acv:" arg
do
    case $arg in
    a) archive=true;;
    c) cached=true;;
    v) version="${OPTARG}";;
    esac
done

if [[ ! -z "$cached" ]]; then #package as cached
    full_path="$full_path-cached"
    args=".bin/packager"
fi

if [[ ! -z "$archive" ]]; then #package as archive
    args="${args} -archive"
fi

if [[ -z "$version" ]]; then #version not provided, use latest git tag
    git_tag=$(git describe --abbrev=0 --tags)
    version=${git_tag:1}
fi

args="${args} -version ${version}"

eval "${args}" "${full_path}"

if [[ -n "$BP_REWRITE_HOST" ]]; then
    sed -i '' -e "s|^uri = \"https:\/\/buildpacks\.cloudfoundry\.org\(.*\)\"$|uri = \"http://$BP_REWRITE_HOST\1\"|g" "$full_path/buildpack.toml"
fi

