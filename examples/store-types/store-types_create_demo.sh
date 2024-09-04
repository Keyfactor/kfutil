#!/usr/bin/env bash
function create_store_type_from_template() {
  echo "Creating store type from template $1.json"
  kfutil store-types templates-fetch | jq -r ."$1" > "$1".json
  kfutil store-types create --from-file "$1".json
}

function download_store_type_template() {
  echo "Downloading store type template $1 to $1.json"
  kfutil store-types templates-fetch | jq -r ."$1" > "$1".json
}

function download_integration_manifest() {
    local repo_name=${1:-kfutil}
    local ref=${2:-main}
    local manifest_file=${3:-integration-manifest.json}
    echo "Downloading integration manifest from Keyfactor/$repo_name@$ref to $manifest_file"
    echo curl -o $manifest_file https://raw.githubusercontent.com/Keyfactor/${repo_name}/${ref}/integration-manifest.json
    curl -o $manifest_file https://raw.githubusercontent.com/Keyfactor/${repo_name}/${ref}/integration-manifest.json
}

function create_store_type_from_manifest() {
    local shortname=$1
    local manifest_file=${2:-integration-manifest.json}

    # check if manifest file exists
    if [ ! -f "$manifest_file" ]; then
        echo "Manifest file '$manifest_file' does not exist"
        return 1
    fi

    # check if $1 is empty
    if [ -z "$shortname" ]; then
        echo "StoreType 'shortname' is required"
        cat $manifest_file | jq '.about.orchestrator.store_types[] | .ShortName'
        return 1
    fi

    echo "Creating store type from manifest $manifest_file for $shortname"
    jq --arg shortname "$shortname" '.about.orchestrator.store_types[] | select(.ShortName == $shortname)' "$manifest_file" > "$shortname".json

    kfutil store-types create --from-file "$shortname".json
}

# Examples
#create_store_type_from_template "BIPCamera" # Use for online creation
function offline_create_store_type_from_template() {
  # Use for offline creation
  local store_type_name=$1
  local orchestrator_name=$2
  local orchestrator_version=${3:-main}
  local manifest_file=${4:-test-manifest.json}
  echo "Downloading store type template $store_type_name"
  download_integration_manifest "${orchestrator_name}" "${orchestrator_version}" "${manifest_file}"

  # Use for offline creation
  echo "Download the latest kfutil binary from https://github.com/Keyfactor/kfutil/releases/latest"
  echo "Copy 'kfutil' and the '${manifest_file}' to offline machine and then run the following command"
  echo create_store_type_from_manifest "${store_type_name}" "${manifest_file}"
  #create_store_type_from_manifest "${store_type_name}" "${manifest_file}" # Uncomment to run directly
}

function offline_create_bipcamera_store_type() {
  offline_create_store_type_from_template "BIPCamera" "bosch-ipcamera-orchestrator" "main" "bipcamera-manifest.json"
}

offline_create_bipcamera_store_type