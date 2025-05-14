# v1.7.0

## Features

### CLI

- `stores import csv`: supports interactive credential input, as well as input via flags and environmental
  variables. [docs](docs/kfutil_stores_import_csv.md)

## Fixes

### CLI

- `stores import csv`: providing a `Password(/StorePassword)` does not crash CLI.
- `stores import csv`: results CSV retains input header ordering.
- `stores import csv`: Handle `BOM` characters in an input CSV file.
- `store-types create`: URL encode `-b` parameter when passed.
- `store-types create`: Initialize logger before fetching store-type definitions.
- `stores rot`: Re-enabled and improved logging.

# v1.6.2

## Fixes

### CLI

- `version`: Correct version is reported for `kfutil version`

# v1.6.1

## Fixes

### CLI

- `auth`: When using `oauth` pass empty list for `scopes` if no scopes are provided, rather than default scope `openid`
- `auth`: Output env and config file errors when both are encountered rather than just config file errors.

## Chores

- `store-types`: Update embedded `store-type` definitions to latest.

# v1.6.0

## Features

### CLI

- `auth`: Added support for authenticating to Keyfactor Command using a oAuth2 client credentials or access token.
- `logout`: Added support for logging out of specific `profile` and `config-file`.
- `logout`: Added `yes|no` prompt for logout actions, which can be skipped by using the `--no-prompt` flag.

### Store Types

- `store-types create`: Added support for creating store types from a local file in `integration-manifest.json` format.
- `store-types create`: Added support for creating store types specified by a Keyfactor repo name and optional branch
  ref.

## Fixes

### CLI

- Fixed an issue where the CLI would sometimes terminate with no error messages when calling the
  `keyfactor-go-client-sdk`
- `auth`: When passing `--config` and/or `--profile` flags, and a failure occurs, the CLI will now return an error
  message
  rather attempt environment variable and default config file/profile fallbacks.

### Stores

- `import csv`: Converts all `int` properties to `string` since Keyfactor Command does not support `int` properties.
- `import csv`: Returns useful error message when invalid `store-type-name` or `store-type-id` are passed rather than
  panic.

## Chores

- `deps`: Bump `go` version to `1.23`.
- `deps`: Bump `azure-sdk-for-go/sdk/azidentity` version to `v1.8.0`.
- `deps`: Bump `AzureAD/microsoft-authentication-library-for-go` to `v1.3.2`.
- `deps`: Bump `keyfactor-go-client-sdk` version to `v2.0.0`.
- `deps`: Bump `keyfactor-go-client` version to `v3.0.0`.
- `deps`: Bump `creack/pty` to `v1.1.24`.
- `deps`: Bump `stretchr/testify` to `v1.10.0`.
- `deps`: Bump `x/crypto` to `v0.30.0`.
- `deps`: Bump `x/term` to `v0.27.0`.
- `deps`: Bump `x/sys` to `v0.28.0`.
- `deps`: Bump `x/text` to `v0.21.0`.

# v1.5.1

## Fixes

- fix(pkg): Bump module version to `v1.5.1` to fix an issue with the `1.5.0` release.

# v1.5.0

## Features

### CLI

- The CLI will now embed the store_type definitions for each release of `kfutil`.
- Add global flag `--offline` to allow for offline mode. This will prevent the CLI from making requests to GitHub for
  store types and store type templates and will use embedded store types and templates instead.

## Fixes

### Stores

- `stores export --all`: Correctly paginates through all stores when exporting.

### CLI

- No longer log before the `--debug` flag is evaluated.

# v1.4.0

## Features

### Stores

- `stores import generate-template`: New sub CLI to generate a CSV template for bulk importing
  stores. [See docs](docs/kfutil_stores_import_generate-template.md)`.
- `stores delete`: Support for user interactive mode.
- `stores delete`: Support of delete from CSV file.
- `stores export`: Supports `--all` flag and user interactive mode

## Fixes

- Various null pointer references when nothing and/or empty inputs/responses are received.
- Installer script checksum check now validates properly. #119
- `stores import` sub CLI is now listed and documented #71

### Store Types

- Empty `storepath` values are no longer passed to the API. #56

### PAM Types

- Handle duplicate provider type that is already created without crashing. #139

## Docs

- [Examples for certificate store bulk operations](https://github.com/Keyfactor/kfutil/tree/epic_54795/examples/cert_stores/bulk_operations#readme)

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

- Added `--output-to-integration-manifest` flag to `kfutil store-types get` to download a remote certificate store type
  definition into an `integration-manifest.json` file locally.
- Updated usage:
  `kfutil store-types get [-i <store-type-id> | -n <store-type-name>] [-b <git-ref>] [-g | --output-to-integration-manifest]`

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
