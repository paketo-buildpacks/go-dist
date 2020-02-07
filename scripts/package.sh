#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=.util/tools.sh
source "${PWD}/scripts/.util/tools.sh"

# shellcheck source=.util/print.sh
source "${PWD}/scripts/.util/print.sh"

function main() {
    util::tools::packager::install --directory "${BUILDPACKDIR}/.bin"

    PACKAGE_DIR=${PACKAGE_DIR:-"$(dirname "${BUILDPACKDIR}")_$(openssl rand -hex 4)"}

    local full_path args version cached archive
    full_path="$(realpath "${PACKAGE_DIR}")"
    args="${BUILDPACKDIR}/.bin/packager -uncached"

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --archive|-a)
          archive="true"
          shift 1
          ;;

        --cached|-c)
          cached="true"
          shift 1
          ;;

        --version|-v)
          version="${2}"
          shift 2
          ;;

        "")
          # skip if the argument is empty
          shift 1
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    if [[ -n "${cached:-}" ]]; then
        full_path="${full_path}-cached"
        args="${BUILDPACKDIR}/.bin/packager"
    fi

    if [[ -n "${archive:-}" ]]; then
        args="${args} -archive"
    fi

    if [[ -z "${version:-}" ]]; then
        version="$(cd "${BUILDPACKDIR}" && git describe --abbrev=0 --tags)"
    fi

    args="${args} -version ${version}"

    pushd "${BUILDPACKDIR}" > /dev/null
        eval "${args}" "${full_path}"
    popd > /dev/null

    if [[ -n "${BP_REWRITE_HOST:-}" ]]; then
        sed -i '' -e "s|^uri = \"https:\/\/buildpacks\.cloudfoundry\.org\(.*\)\"$|uri = \"http://${BP_REWRITE_HOST}\1\"|g" "${full_path}/buildpack.toml"
    fi

}

main "${@:-}"
