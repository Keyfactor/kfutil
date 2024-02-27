# Keyfactor Command Utility (kfutil)

`kfutil` is a go-lang CLI wrapper for Keyfactor Command API. It also includes other utility/helper functions around automating common Keyfactor Command operations.

#### Integration status: Production - Ready for use in production environments.

<!-- toc -->

- [About the Keyfactor API Client](#about-the-keyfactor-api-client)
- [Support for Keyfactor Command Utility (kfutil)](#support-for-keyfactor-command-utility-kfutil)
- [Quickstart](#quickstart)
  * [Linux/MacOS](#linuxmacos)
    + [Prerequisites:](#prerequisites)
    + [Installation:](#installation)
  * [Windows](#windows)
    + [Prerequisites:](#prerequisites-1)
    + [Installation:](#installation-1)
- [Environmental Variables](#environmental-variables)
  * [Linux/MacOS:](#linuxmacos)
  * [Windows Powershell:](#windows-powershell)
- [Authentication Providers](#authentication-providers)
- [Commands](#commands)
  * [Login](#login)
  * [Logout](#logout)
- [Commands](#commands-1)
  * [Bulk operations](#bulk-operations)
    + [Bulk create cert stores](#bulk-create-cert-stores)
    + [Bulk create cert store types](#bulk-create-cert-store-types)
  * [Root of Trust](#root-of-trust)
  * [Root of Trust Quickstart](#root-of-trust-quickstart)
    + [Generate Certificate List Template](#generate-certificate-list-template)
    + [Generate Certificate Store List Template](#generate-certificate-store-list-template)
    + [Run Root of Trust Audit](#run-root-of-trust-audit)
    + [Run Root of Trust Reconcile](#run-root-of-trust-reconcile)
  * [Certificate Store Inventory](#certificate-store-inventory)
    + [Show the inventory of a certificate store](#show-the-inventory-of-a-certificate-store)
    + [Add certificates to certificate stores](#add-certificates-to-certificate-stores)
    + [Remove certificates from certificate stores](#remove-certificates-from-certificate-stores)
- [Development](#development)
  * [Adding a new command](#adding-a-new-command)
<!-- tocstop -->

## About the Keyfactor API Client

This API client allows for programmatic management of Keyfactor resources.

## Support for Keyfactor Command Utility (kfutil)

Keyfactor Command Utility (kfutil) is open source and supported on best effort level for this tool/library/client.  This means customers can report Bugs, Feature Requests, Documentation amendment or questions as well as requests for customer information required for setup that needs Keyfactor access to obtain. Such requests do not follow normal SLA commitments for response or resolution. If you have a support issue, please open a support ticket via the Keyfactor Support Portal at https://support.keyfactor.com/

[!NOTE] To report a problem or suggest a new feature, use the **[Issues](../../issues)** tab. If you want to contribute actual bug fixes or proposed enhancements, use the **[Pull requests](../../pulls)** tab.

## Quickstart

### Linux/MacOS
#### Prerequisites:
- [jq](https://stedolan.github.io/jq/download/) CLI tool, used to parse JSON output.
- Either
  - [curl](https://curl.se/download.html) CLI tool, used to download the release files.
  - OR [wget](https://www.gnu.org/software/wget/) CLI tool, used to download the release files.
- [unzip](https://linuxize.com/post/how-to-unzip-files-in-linux/#installing-unzip) CLI tool, used to unzip the release
- [openssl](https://www.openssl.org/source/) CLI tool, used to validate package checksum.
- `$HOME/.local/bin` in your `$PATH` and exists if not running as root, else `/usr/local/bin` if running as root.

#### Installation:
```bash
bash <(curl -s https://raw.githubusercontent.com/Keyfactor/kfutil/main/install.sh)
````

### Windows
#### Prerequisites:
- Powershell 5.1 or later

#### Installation:
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/Keyfactor/kfutil/main/install.ps1" -OutFile "install.ps1"
# Install kfutil to $HOME/AppData/Local/Microsoft/WindowsApps.
# Use Get-Help .\install.ps1 -Full for help and examples.
.\install.ps1
```

## Environmental Variables

All the variables listed below need to be set in your environment. The `kfutil` command will look for these variables
and use them if they are set. If they are not set, the utility will fail to connect to Keyfactor.

| Variable Name      | Description                                                                              |
|--------------------|------------------------------------------------------------------------------------------|
| KEYFACTOR_HOSTNAME | The hostname of your Keyfactor instance. ex: `my.domain.com`                             |
| KEYFACTOR_USERNAME | The username to use to connect to Keyfactor. Do not include the domain. ex: `myusername` |
| KEYFACTOR_PASSWORD | The password to use to connect to Keyfactor. ex: `mypassword`                            |
| KEYFACTOR_DOMAIN   | The domain to use to connect to Keyfactor. ex: `mydomain`                                |
| KEYFACTOR_API_PATH | The path to the Keyfactor API. Defaults to `/KeyfactorAPI`.                              |
| KFUTIL_EXP         | Set to `1` or `true` to enable experimental features.                                    |
| KFUTIL_DEBUG       | Set to `1` or `true` to enable debug logging.                                            |

### Linux/MacOS:

```bash
export KEYFACTOR_HOSTNAME="<mykeyfactorhost.mydomain.com>"
export KEYFACTOR_USERNAME="<myusername>" # Do not include domain
export KEYFACTOR_PASSWORD="<mypassword>"
export KEYFACTOR_DOMAIN="<mykeyfactordomain>"
```

Additional variables:

```bash
export KEYFACTOR_API_PATH="/KeyfactorAPI" # Defaults to /KeyfactorAPI if not set ex. my.domain.com/KeyfactorAPI
export KFUTIL_EXP=0 # Set to 1 or true to enable experimental features
export KFUTIL_DEBUG=0 # Set to 1 or true to enable debug logging
```

### Windows Powershell:

```powershell
$env:KEYFACTOR_HOSTNAME = "<mykeyfactorhost.mydomain.com>"
$env:KEYFACTOR_USERNAME = "<myusername>" # Do not include domain
$env:KEYFACTOR_PASSWORD = "<mypassword>"
$env:KEYFACTOR_DOMAIN = "<mykeyfactordomain>"
```

Additional variables:

```bash
$env:KEYFACTOR_API_PATH="/KeyfactorAPI" # Defaults to /KeyfactorAPI if not set ex. my.domain.com/KeyfactorAPI
$env:KFUTIL_EXP=0 # Set to 1 or true to enable experimental features
$env:KFUTIL_DEBUG=0 # Set to 1 or true to enable debug logging
```

## Authentication Providers

`kfutil` supports the following authentication providers in order of precedence:

| Provider Type                | Description                                                                                                                                                             |
|------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Azure Key Vault via Azure ID | This provider will read the Keyfactor Command credentials from Azure Key Vault. For more info review the [auth providers](docs/auth_providers.md#azure-key-vault) docs. |
| Environment                  | This provider will read the Keyfactor Command credentials from the environment variables listed above.                                                                  |
| File                         | This is the default provider. It will read the credentials from a file on disk at `$HOME/.keyfactor/command_config.json`                                                |
| User Interactive             | This provider will prompt the user for their credentials.                                                                                                               |

## Commands

### Login

For full documentation on the `login` command, see the [login](docs/kfutil_login.md) documentation.

*WARNING* - The `login` command will store your Keyfactor credentials in a file on your local machine. This file is not
encrypted and is not secure. It is recommended that you use the `login` command only on your local machine and not on a
shared machine. Instead of using the `login` command, you can set the environmental variables listed above.

```bash
kfutil login
```

### Logout

For full documentation on the `logout` command, see the [logout](docs/kfutil_logout.md) documentation.

*WARNING* - This will delete the file containing your Keyfactor credentials at `$HOME/.keyfactor/command_config.json`.

```bash
kfutil logout
```

## Commands

### Bulk operations

#### Bulk create cert stores

For full documentation, see [stores import](docs/kfutil_stores_import.md).

This will attempt to process a CSV input file of certificate stores to create. The template can be generated by
running: `kfutil stores import generate-template` command.

```bash
kfutil stores import csv --file <file name to import>
```

```bash
kfutil stores import --help       
Tools for generating import templates and importing certificate stores

Usage:
  kfutil stores import [command]

Available Commands:
  csv               Create certificate stores from CSV file.
  generate-template For generating a CSV template with headers for bulk store creation.

Flags:
  -h, --help   help for import

Global Flags:
      --api-path string                API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI) (default "KeyfactorAPI")
      --auth-provider-profile string   The profile to use defined in the securely stored config. If not specified the config named 'default' will be used if it exists. (default "default")
      --auth-provider-type string      Provider type choices: (azid)
      --config string                  Full path to config file in JSON format. (default is $HOME/.keyfactor/command_config.json)
      --debug                          Enable debugFlag logging.
      --domain string                  Domain to use for authenticating to Keyfactor Command.
      --exp                            Enable expEnabled features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)
      --format text                    How to format the CLI output. Currently only text is supported. (default "text")
      --hostname string                Hostname to use for authenticating to Keyfactor Command.
      --no-prompt                      Do not prompt for any user input and assume defaults or environmental variables are set.
      --password string                Password to use for authenticating to Keyfactor Command. WARNING: Remember to delete your console history if providing kfcPassword here in plain text.
      --profile string                 Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.
      --username string                Username to use for authenticating to Keyfactor Command.

Use "kfutil stores import [command] --help" for more information about a command.
```

#### Bulk create cert store types

For full documentation, see [store-types](docs/kfutil_store-types.md).

This will attempt to process a CSV input file of certificate store types to create. The template can be generated by
running: `kfutil generate-template --type bulk-certstore-types` command.

```bash
kfutil store-types create --name $STORE_TYPE_NAME
```

```bash
kfutil store-types --help             
A collections of APIs and utilities for interacting with Keyfactor Command certificate store types.

Usage:
  kfutil store-types [command]

Available Commands:
  create          Create a new certificate store type in Keyfactor Command.
  delete          Delete a specific store type by ID.
  get             Get a specific store type by either name or ID.
  list            List certificate store types.
  templates-fetch Fetches store type templates from Keyfactor's Github.
  update          Update a certificate store type in Keyfactor.

Flags:
  -h, --help   help for store-types

Use "kfutil store-types [command] --help" for more information about a command.
```

### Root of Trust

For full documentation, see [stores rot](docs/kfutil_stores_rot.md).

The root of trust (rot) utility is a tool that allows you to bulk manage Keyfactor certificate stores and ensure that a
set of defined certificates are present in each store that meets a certain set of criteria or no criteria at all.

### Root of Trust Quickstart

```bash
echo "Generating cert template file certs_template.csv"
kfutil stores rot generate-template-rot --type certs
# edit the certs_template.csv file
echo "Generating stores template file stores_template.csv"
kfutil stores rot generate-template-rot --type stores
# edit the stores_template.csv file
kfutil stores rot audit --add-certs certs_template.csv --stores stores_template.csv #This will audit the stores and generate a report file
# review/edit the report file generated `rot_audit.csv`
kfutil stores rot reconcile --import-csv
# Alternatively this can be done in one step
kfutil stores rot reconcile --add-certs certs_template.csv --stores stores_template.csv
```

#### Generate Certificate List Template

For full documentation, see [stores rot generate template](docs/kfutil_stores_rot_generate-template.md).

This will write the file `certs_template.csv` to the current directory.

```bash
kfutil stores generate-template-rot --type certs
```

#### Generate Certificate Store List Template

For full documentation, see [stores rot generate template](docs/kfutil_stores_rot_generate-template.md).

This will write the file `stores_template.csv` to the current directory. For full documentation

```bash
kfutil stores generate-template-rot --type stores
```

#### Run Root of Trust Audit

For full documentation, see [stores rot audit](docs/kfutil_stores_rot_audit.md).

Audit will take in a list of certificates and a list of certificate stores and check that the certificate store's
inventory either contains the certificate or does not contain the certificate based on the `--add-certs` and
`--remove-certs` flags. These flags can be used together or separately. The aforementioned flags take in a path to CSV
files containing a list of certificate thumbprints. To generate a template for these files, run the following command:

```bash
kfutil stores rot generate-template --type certs
```

To prepopulate the template file you can provide `--cn` multiple times.

```bash
kfutil stores rot generate-template --type certs \
  --cn <cert subject name> \
  --cn <additional cert subject name>
```

In addition, you must provide a list of stores you wish to audit. To generate a template for this file, run the
following
command:

```bash
kfutil stores rot generate-template --type stores
```

To prepopulate the template file you can provide `--store-type` and `--container-type` multiple times.

```bash
kfutil stores rot generate-template --type stores \
  --store-type <store type name> \
  --store-type <additional store type name> \
  --container-type <container type name> \
  --container-type <additional container type name>
```

With all the files generated and populated, you can now run the audit command:

```bash
kfutil stores rot audit \
  --stores stores_template.csv \
  --add-certs certs_template.csv \
  --remove-certs certs_template2.csv
```

This will generate an audit file that contains the results of the audit and actions that will be taken if `reconcile` is
executed. By default, the audit file will be named `rot_audit.csv` and will be written to the current directory. To
output
the audit file to a different location, use the `--output` flag:

```bash
kfutil stores rot audit \
  --stores stores.csv \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --output /path/to/output/autdit_file.csv
```

#### Run Root of Trust Reconcile

For full documentation, see [stores rot](docs/kfutil_stores_rot_reconcile.md).

Reconcile will take in a list of certificates and a list of certificate stores and check that the certificate store's
inventory either contains the certificate or does not contain the certificate based on the `--add-certs` and
`--remove-certs` flags. These flags can be used together or separately. The aforementioned flags take in a path to CSV
files containing a list of certificate thumbprints. To generate a template for these files, run the following command:

```bash
kfutil stores rot generate-template --type certs
```

To pre-populate the template file you can provide `--cn` multiple times.

```bash
kfutil stores rot generate-template --type certs \
  --cn <cert subject name> \
  --cn <additional cert subject name>
```

In addition, you must provide a list of stores you wish to reconcile. To generate a template for this file, run the
following
command:

```bash
kfutil stores rot generate-template --type stores
```

To pre-populate the stores template file you can provide multiple values in any combination of the following flags:

```bash
kfutil stores rot generate-template --type stores \
  --store-type <store type name> \
  --store-type <additional store type name> \
  --container-type <container type name> \
  --container-type <additional container type name>
```

With all the files generated and populated, you can now run the reconcile command:

```bash
kfutil stores rot reconcile \
  --stores stores_template.csv \
  --add-certs certs_template.csv \
  --remove-certs certs_template2.csv
```

This will generate an audit file that contains the results of the audit and actions will immediately execute those
actions.
By default, the reconcile file will be named `rot_audit.csv` and will be written to the current directory. To output
the reconcile file to a different location, use the `--output` flag:

```bash
kfutil stores rot reconcile \
  --stores stores.csv \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --output /path/to/output/audit_file.csv
```

Alternatively you can provide an audit CSV file as an input to the reconcile command using the `--import-csv` flag:

```bash
kfutil stores rot reconcile \
  --import-csv /path/to/audit_file.csv
```

### Certificate Store Inventory

For full documentation, see [stores inventory](docs/kfutil_stores_inventory.md).

#### Show the inventory of a certificate store

For full documentation, see [stores inventory show](docs/kfutil_stores_inventory_show.md).

```bash
# Show by store ID:
```bash
kfutil stores inventory show --sid <store id>

# Nested command lookup: shows inventory of first cert store found
kfutil stores inventory show \
  --sid $(kfutil stores list | jq -r ".[0].Id")
```

Show by client machine name:

```bash
kfutil stores inventory show --client <machine name>

# Nested command lookup: shows inventory of first cert store found
kfutil stores inventory show \
  --client $(kfutil orchs list | jq -r ".[0].ClientMachine")
```

#### Add certificates to certificate stores

For full documentation, see [stores inventory add](docs/kfutil_stores_inventory_add.md).

```bash
# Add 2 certs to 2 certificate stores
kfutil stores inventory add \
  --sid <store id> \
  --sid <additional store id> \
  --cn <cert subject name> \
  --cn <additional cert subject name>
```

#### Remove certificates from certificate stores

For full documentation, see [stores inventory remove](docs/kfutil_stores_inventory_remove.md).

```bash
# Remove 2 certs from all stores associated with a client machine
kfutil stores inventory remove \
  --client <machine name> \
  --cn <cert subject name> \
  --cn <additional cert subject name>
```

## Development

This CLI developed using [cobra](https://umarcor.github.io/cobra/)

### Adding a new command

```bash
cobra-cli add <my-new-command>
```

alternatively you can specify the parent command

```bash
cobra-cli add <my-new-command> -p '<parent>Cmd'
```

