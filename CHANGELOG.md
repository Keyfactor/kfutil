# v1.2.0
## Features
feat(auth): Added support for sourcing credentials from [Azure Key Vault using Azure ID](docs/auth_providers.md#azure-key-vault)
feat(cli): Added enhanced logging when `KFUTIL_DEBUG` is set.
feat(store-types): `store-types create` now supports the `--all` flag.
feat(stores): `stores` sub CLI is now non-experimental. [See docs](docs/kfutil_stores.md)
feat(stores): Bulk import of stores via `stores import csv` sub CLI is now non-experimental. [See docs](docs/kfutil_stores_import_csv.md)
feat(stores): Added `delete` command to stores as well as a `--all` option.

## Bug Fixes
fix(login): the default `APIPath` no longer overwrites preexisting values.

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