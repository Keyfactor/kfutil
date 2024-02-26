# v1.4.0
## Features
- `stores import generate-template`: Sub CLI is now non-experimental. Generate a CSV template for bulk importing certificate stores. [See docs](docs/kfutil_stores_import_generate-template.md)`

## Fixes
- Various null pointer references when nothing and/or empty inputs/responses are received.
- Installer script checksum check now validates properly. #119
- `stores import` sub CLI is now listed and documented #71

### Store Types
- Empty `storepath` values are no longer passed to the API.  #56

### PAM Types
- Handle duplicate provider type that is already created without crashing. #139

# v1.3.2

### Package
- Bump deps `cobra` version to `v1.8.0`, `azcore` version to `v1.9.0`, `pty` version to `v1.1.21`

# v1.3.1

## Bug Fixes

### Package
- Bump package version to `1.3.1` to fix an issue with the `1.3.0` release.

### Installer
- Remove `v` prefix from installer URL path to accommodate for the new build process.

# v1.3.0
## Features

### StoreTypes
- Added `--output-to-integration-manifest` flag to `kfutil store-types get` to download a remote certificate store type definition into an `integration-manifest.json` file locally.
- Updated usage: `kfutil store-types get [-i <store-type-id> | -n <store-type-name>] [-b <git-ref>] [-g | --output-to-integration-manifest]`

# v1.2.1

## Bug Fixes

### StoreTypes

- `store-type templates-fetch` now supports a `--git-ref` flag to specify a specific branch, tag, or commit to fetch
  templates from.
- `store-types create` now omits the `StorePath` value when not specified. This fixes the issue where the `StorePath`
  value was being set to "" which Command interpreted as only allowing "" for store paths on created store types.

### CLI

- `login` now un-hidden from CLI help.

# v1.2.0

## Features
 
### Auth

- Added support for sourcing credentials from [Azure Key Vault using Azure ID](docs/auth_providers.md#azure-key-vault)

### CLI

- Added enhanced logging when `KFUTIL_DEBUG` is set.

### Helm

- `helm uo` New sub CLI to configure UO Helm Chart. [See docs](docs/kfutil_helm_uo)

### Orchestrator Extensions

- `orchs ext`: New sub CLI to download orchestrator extensions from GitHub. [See docs](docs/kfutil_orchs_ext)

### Stores

- `stores`: Sub CLI is now non-experimental. [See docs](docs/kfutil_stores.md)
- `stores import csv`: Bulk import of stores via CSV is now
  non-experimental. [See docs](docs/kfutil_stores_import_csv.md)
- `stores delete`: Added delete a store from Keyfactor Command, as well as a `--all` option that will delete all stores
  from Keyfactor Command.

### StoreTypes

- `store-types create`: now supports the `--all` flag and will attempt to create all store types available from
  Keyfactor's GitHub org.

## Bug fixes

### Auth

- the default `APIPath` no longer overwrites preexisting values.

# v1.1.0

## Features

- `pam`: [kfutil pam](docs/kfutil_pam.md)

# v1.0.0

## Overview

Initial release of the Keyfactor Command Utility (kfutil)

Production Supported CLIs:

- `login`: [kfutil login](docs/kfutil_login.md)
- `store-types`: [kfutil store-types](docs/kfutil_store-types.md)
- `stores rot`: [kfutil rot](docs/kfutil_stores_rot.md)
