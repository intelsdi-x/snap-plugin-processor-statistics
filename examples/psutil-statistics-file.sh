#!/bin/bash


set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# source the common bash script 
. "${__dir}/../scripts/common.sh"

# ensure PLUGIN_PATH is set
TMPDIR=${TMPDIR:-"/tmp"}
PLUGIN_PATH=${PLUGIN_PATH:-"${TMPDIR}/snap/plugins"}
mkdir -p $PLUGIN_PATH

_info "Get latest plugins"
(cd $PLUGIN_PATH && curl -fL -sSO http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest_build/linux/x86_64/snap-plugin-publisher-file && chmod 755 snap-plugin-publisher-file)
(cd $PLUGIN_PATH && curl -fL -sSO http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-psutil/latest/linux/x86_64/snap-plugin-collector-psutil && chmod 755 snap-plugin-collector-psutil)
(cd $PLUGIN_PATH && curl -fL -sSO http://snap.ci.snap-telemetry.io/plugins/snap-plugin-processor-statistics/latest_build/linux/x86_64/snap-plugin-processor-statistics && chmod 755 snap-plugin-processor-statistics)

sleep 10
_info "loading plugins"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-collector-psutil"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-publisher-file"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-processor-statistics"

_info "creating and starting a task"
snapctl task create -t "${__dir}/tasks/psutil-statistics-file.json"
