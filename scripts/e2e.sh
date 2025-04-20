#!/usr/bin/env bash

source scripts/utils.sh

trap "fail_trap" EXIT
make PROVIDER_TOKEN=$PROVIDER_TOKEN WORKFLOW_ACCESS_TOKEN=$WORKFLOW_ACCESS_TOKEN local.build
installFileFromLocal
echo
echo "helm-kanvas-snapshot is installed at ${HELM_PLUGIN_DIR}/bin/helm-kanvas-snapshot"
echo
echo "See https://github.com/$PROJECT_GH#readme for more information on getting started."