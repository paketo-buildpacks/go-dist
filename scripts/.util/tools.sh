#!/usr/bin/env bash
set -eu
set -o pipefail

# shellcheck source=./print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

# shellcheck source=./git.sh
source "$(dirname "${BASH_SOURCE[0]}")/git.sh"

function util::tools::install() {
    util::print::title "Installing Testing Tools"

    local dir pack_version cnb2cf_version
    cnb2cf_version=""

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --help|-h)
          util::tools::usage
          exit 0
          ;;

        --directory)
          dir="${2}"
          shift 2
          ;;

        --pack-version)
          pack_version="${2}"
          shift 2
          ;;

        --cnb2cf-version)
          cnb2cf_version="${2}"
          shift 2
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    mkdir -p "${dir}"

    util::tools::pack::install "${dir}" "${pack_version}"
    util::tools::packager::install "${dir}"
    util::tools::cnb2cf::install "${dir}" "${cnb2cf_version}"
}

function util::tools::pack::install() {
    local dir version os
    dir="${1}"
    version="${2}"

    util::print::title "Installing pack"

    os="$(uname -s)"

    if [[ "${os}" == "Darwin" ]]; then
        os="macos"
    elif [[ "${os}" == "Linux" ]]; then
        os="linux"
    else
        util::print::error "Unsupported operating system"
    fi

    if [[ ! -f "${dir}/pack" ]]; then
        util::print::info "--> installing..."
    elif [[ "$("${dir}/pack" version | cut -d ' ' -f 1)" != *${version}* ]]; then
        rm "${dir}/pack"
        util::print::info "--> updating..."
    else
        util::print::info "--> skipping..."
        return 0
    fi

    GIT_TOKEN="$(util::git::token::fetch)"

    if [[ "${version}" == "latest" ]]; then
        local url
        if [[ "${os}" == "macos" ]]; then
            url="$(
                curl -s \
                    -H "Authorization: token ${GIT_TOKEN}" \
                    https://api.github.com/repos/buildpacks/pack/releases/latest \
                    | jq --raw-output '.assets[1] | .browser_download_url'
                )"
        else
            url="$(
                curl -s \
                    -H "Authorization: token ${GIT_TOKEN}" \
                    https://api.github.com/repos/buildpacks/pack/releases/latest \
                    | jq --raw-output '.assets[0] | .browser_download_url'
                )"
        fi
        util::tools::pack::expand "${dir}" "${url}"
    else
        local tarball
        tarball="pack-${version}-${os}.tar.gz"
        url="https://github.com/buildpacks/pack/releases/download/v${version}/${tarball}"
        util::tools::pack::expand "${dir}" "${url}"
    fi
}

function util::tools::pack::expand() {
    local dir url version
    dir="${1}"
    url="${2}"
    tarball="$(echo "${url}" | sed "s/.*\///")"
    version="v$(echo "${url}" | sed 's/pack-//' | sed 's/-.*//')"

    wget -q "${url}"
    tar xzf "${tarball}" -C "${dir}"
    rm "${tarball}"
}

function util::tools::packager::install () {
    local dir
    dir="${1}"

    if [[ ! -f "${dir}/packager" ]]; then
        util::print::title "Installing packager"
        go build -o "${dir}/packager" github.com/cloudfoundry/libcfbuildpack/packager
    fi
}

function util::tools::cnb2cf::install() {
    local dir version
    dir="${1}"
    version="${2}"

    if [[ -n "${version}" && ! -f "${dir}/cnb2cf" ]]; then
        util::print::title "Installing cnb2cf"
        go build -o "${dir}/cnb2cf" github.com/cloudfoundry/cnb2cf
    fi
}
