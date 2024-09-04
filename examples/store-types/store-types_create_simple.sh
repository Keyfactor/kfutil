#!/usr/bin/env bash
function create_store_type_from_template() {
  kfutil store-types templates-fetch | jq -r ."$1" > "$1".json
  kfutil store-types create --from-file "$1".json
}

function create_store_type_from_manifest() {
    local shortname=$1
    local manifest_file=${2:-integration-manifest.json}

    jq --arg shortname "$shortname" '.about.orchestrator.store_types[] | select(.ShortName == $shortname)' "$manifest_file" > "$shortname".json

    kfutil store-types create --from-file "$shortname".json
}

# Examples
create_store_type_from_template "AKV"
create_store_type_from_manifest "K8SSecret" test-manifest.json