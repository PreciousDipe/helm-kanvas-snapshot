#!/usr/bin/env bash

source scripts/utils.sh


# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e
make build
installFile
echo
echo "helm-kanvas-snapshot is installed at ${HELM_PLUGIN_DIR}/bin/helm-kanvas-snapshot"
echo
echo "See https://github.com/$PROJECT_GH#readme for more information on getting started."

