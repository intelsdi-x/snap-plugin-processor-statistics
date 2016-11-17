#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename $__proj_dir)"

. "${__dir}/common.sh"

# NOTE: these variables control the docker-compose image.
export PLUGIN_SRC="${__proj_dir}"
export PROJECT_NAME="${__proj_name}"

TEST_TYPE="${TEST_TYPE:-"large"}"

scripts_folder="${__proj_dir}/scripts"

_spec() {
  (cd "${scripts_folder}" && "$@")
}

_info "spec folder : $scripts_folder"

[[ -f "${__proj_dir}/build/linux/x86_64/snap-plugin-processor-statistics" ]] || (cd "${__proj_dir}" && make)

# NOTE: we need to copy the file because how docker build works
cp "${__proj_dir}/build/linux/x86_64/snap-plugin-processor-statistics" "${__proj_dir}/examples/snap-plugin-processor-statistics"

[[ -d "${scripts_folder}/.bundle" ]] || _spec bundle install

_debug "running test: ${TEST_TYPE}"
_spec bundle exec rspec ./spec/task_spec.rb

rm "${__proj_dir}/examples/snap-plugin-processor-statistics"
