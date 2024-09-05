# store-types examples

This directory contains examples of how to use `kfutil` to manage Keyfactor Command store types. For exhaustive details
on the `store-types` command, see the [cli docs](../../docs/kfutil_store-types.md).

- [store-types examples](#store-types-examples)
   - [Examples](#examples)
      - [create](#create)
         - [User Interactive](#user-interactive)
         - [Non-Interactive](#non-interactive)
            - [From File](#from-file)
               - [Simple](#simple)
               - [Complex](#complex)
   - [Demo Scenarios](#demo-scenarios)
      - [Create Bosch Camera Store Type Offline](#create-bosch-camera-store-type-offline)
         - [Summary](#summary)

## Examples

### create

#### User Interactive

Below is an example of creating a store type using the user interactive mode using `kfutil store-types create` command.

*NOTE*: The list of options is pulled from [this file](../../store_types.json).

```text
kfutil store-types create      
? Choose an option:  [Use arrows to move, type to filter]
> AKV
  AWS-ACM
  Akamai
  AppGwBin
  AzureApp
  AzureAppGw
  AzureSP
Certificate store type AKV created with ID: 150
```

#### Non-Interactive

Below is an example of creating a store type using the non-interactive mode using
`kfutil store-types create $KF_ST_SHORTNAME` command.
*NOTE*: This will pull the latest store type definition from the [kfutil store-types.json](../../store_types.json) file.
```bash
KF_ST_SHORTNAME=AKV
kfutil store-types create $KF_ST_SHORTNAME
```

##### From File
Below is an example of creating a store type using the non-interactive mode using
`kfutil store-types create --file $KF_ST_FILE` command.

###### Simple

```bash
KF_ST_FILE=AKV.json
kfutil store-types create --file $KF_ST_FILE
```

###### Complex

Below is a bit more complex example of creating a store type using the non-interactive mode using a downloaded Command
store type definition JSON file. This file can either be sourced from GitHub using `kfutil store-types templates-fetch`
or from a downloaded `intergration-manifest.json` from a (Keyfactor Universal Orchestrator extension)
[https://github.com/search?q=topic%3Akeyfactor-universal-orchestrator+org%3AKeyfactor+fork%3Atrue&type=repositories].

```bash
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
echo "Uses store-types templates-fetch to get the AKV template from GitHub"
create_store_type_from_template "AKV"

echo "Assumes you have an integration-manifest.json file in the current directory"
create_store_type_from_manifest "AKV" # "path/to/integration-manifest.json" # (Optional) will default to looking for integration-manifest.json in the current directory
```

## Demo Scenarios

### Create Bosch Camera Store Type Offline

#### Summary

This scenario demonstrates how to create a store type for a Bosch Camera offline using a downloaded Command store type
definition JSON file.

#### Steps
1. From an online machine download the latest version of [kfutil](https://github.com/Keyfactor/kfutil/releases/latest)
2. Download the `integration-manifest.json` from
   the [Keyfactor Universal Orchestrator extension](https://github.com/Keyfactor/bosch-ipcamera-orchestrator/blob/main/integration-manifest.json)
   , or use `store-types templates-fetch` to get the latest templates from GitHub.  
BASH:
```bash
kfutil store-types templates-fetch | jq -r ."BIPCamera" > "BIPCamera.json"
```
PowerShell:
```powershell
$kfutilResult = kfutil store-types templates-fetch
$parsedJson = $kfutilResult | ConvertFrom-Json
$parsedJson.BIPCamera | ConvertTo-Json | Set-Content -Path "BIPCamera.json"
```

3. Copy the `kfutil` and `integration-manifest.json`/`BIPCamera.json` files to an offline machine.
4. If using Pull the store type definition from the `integration-manifest.json` file either manually or using

BASH:
```bash
jq --arg shortname \
  "BIPCamera" '.about.orchestrator.store_types[] | select(.ShortName == BIPCamera)' \
  integration-manifest.json > "BIPCamera.json" 
```

PowerShell:
```powershell
# Read the JSON content from the file
$jsonContent = Get-Content -Path "integration-manifest.json" -Raw | ConvertFrom-Json

# Define the short name
$shortName = "BIPCamera"

# Filter the JSON data based on the condition
$filteredResult = $jsonContent.about.orchestrator.store_types | Where-Object { $_.ShortName -eq $shortName }

# Convert the filtered result back to JSON and save it to a file
$filteredResult | ConvertTo-Json -Depth 10 | Set-Content -Path "BIPCamera.json"
```

5. Create the store type using the `kfutil store-types create --from-file BIPCamera.json` command.
