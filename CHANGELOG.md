# v1.3.0
## Features

### Store-Type Get
- Added `--output-to-integration-manifest` flag to `kfutil store-types get` to download a remote certificate store type definition into an `integration-manifest.json` file locally.
  - This path now has the following usage: `get [-i <store-type-id> | -n <store-type-name>] [-b <git-ref>] [-g | --output-to-integration-manifest]`

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
- `stores import csv`: Bulk import of stores via CSV is now non-experimental. [See docs](docs/kfutil_stores_import_csv.md)
- `stores delete`: Added delete a store from Keyfactor Command, as well as a `--all` option that will delete all stores from Keyfactor Command.

### StoreTypes
- `store-types create`: now supports the `--all` flag and will attempt to create all store types available from Keyfactor's GitHub org.

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
