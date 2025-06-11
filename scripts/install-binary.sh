# // Copyright Meshery Authors
# //
# // Licensed under the Apache License, Version 2.0 (the "License");
# // you may not use this file except in compliance with the License.
# // You may obtain a copy of the License at
# //
# //     http://www.apache.org/licenses/LICENSE-2.0
# //
# // Unless required by applicable law or agreed to in writing, software
# // distributed under the License is distributed on an "AS IS" BASIS,
# // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# // See the License for the specific language governing permissions and
# // limitations under the License.

#!/usr/bin/env bash

source ${HELM_PLUGIN_DIR}/bin/helm-kanvas-snapshot/scripts/utils.sh



# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e
initArch
initOS
verifySupported
getDownloadURL
downloadFile
installFileFromZip
echo
echo "helm-kanvas-snapshot is installed at ${HELM_PLUGIN_DIR}/bin/helm-kanvas-snapshot"
echo
echo "See https://github.com/$PROJECT_GH#readme for more information on getting started."

