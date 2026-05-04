# Migrate Static Store Credentials To A PAM Provider

Use this workflow when existing certificate stores have static Keyfactor-encrypted credential values and you want those stores to reference a Keyfactor PAM provider instead.

This is a specialized bulk certificate store update. The workflow uses exported CSV files, edits the `Properties.ServerPassword` credential columns, then syncs the changes back to Keyfactor Command.

## Contents

- [Before You Begin](#before-you-begin)
- [Step 1: Export Stores](#step-1-export-stores)
- [Step 2: Identify The PAM Provider Columns](#step-2-identify-the-pam-provider-columns)
- [Step 3: Build The Sync CSV](#step-3-build-the-sync-csv)
- [RFPKCS12 Examples By PAM Type](#rfpkcs12-examples-by-pam-type)
- [Step 4: Sync The Migration](#step-4-sync-the-migration)
- [Step 5: Verify The Migration](#step-5-verify-the-migration)
- [Notes](#notes)
- [Related Commands](#related-commands)

## Before You Begin

You need:

- `kfutil` configured to authenticate to Keyfactor Command.
- Permission to export and update certificate stores.
- A configured PAM provider in Keyfactor Command.
- The PAM provider ID and any provider parameter names and values required by that provider.
- The target store type short name or store type ID.

Keep each CSV scoped to one certificate store type. The import command accepts one `--store-type-name` or `--store-type-id` per run.

## Step 1: Export Stores

Export the stores you want to migrate:

```bash
kfutil stores export --store-type-name K8SCluster
```

For all store types:

```bash
kfutil stores export --all
```

The export includes the `Id` column required for sync updates.

## Step 2: Identify The PAM Provider Columns

If you already have a store using the target PAM provider, export that store type and use its columns as the pattern.

For a PAM-backed `ServerPassword`, the CSV uses columns like:

```text
Properties.ServerPassword.Provider
Properties.ServerPassword.Parameters.SecretName
Properties.ServerPassword.Parameters.SecretType
Properties.ServerPassword.Parameters.StaticSecretFieldName
```

Example values:

```text
Properties.ServerPassword.Provider=30
Properties.ServerPassword.Parameters.SecretName=dev/aks/kf-integrations
Properties.ServerPassword.Parameters.SecretType=static_json
Properties.ServerPassword.Parameters.StaticSecretFieldName=" "
```

The parameter names depend on the PAM provider type. Use the names exported from a known-good store or from the PAM provider type definition.

## Step 3: Build The Sync CSV

For each row you want to migrate:

- Preserve `Id`.
- Preserve `ClientMachine`, `StorePath`, `AgentId`, and other store configuration values.
- Add the PAM provider columns if they are not already present.
- Set `Properties.ServerPassword.Provider` to the PAM provider ID.
- Set the `Properties.ServerPassword.Parameters.*` columns to the provider parameter values.
- Leave `Properties.ServerPassword.SecretValue` empty if that column exists.

Example:

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretName,Properties.ServerPassword.Parameters.SecretType,Properties.ServerPassword.Parameters.StaticSecretFieldName,Properties.ServerPassword.SecretValue
13b0b2c5-eb27-4885-91ec-fad35d0268df,kf-integrations,fresh,30,dev/aks/kf-integrations,static_json," ",
```

Do not put the masked export value `********************` into a new direct secret value column. That is a placeholder, not the original secret.

## RFPKCS12 Examples By PAM Type

The embedded store type short name is `RFPkcs12`; use that exact value with `--store-type-name`.

These examples show the columns to migrate an `RFPkcs12` row from static values to PAM-backed `Properties.ServerPassword` and PAM-backed store `Password`. Replace provider IDs, store IDs, paths, and PAM parameter values with values from your environment.

If you are migrating `Properties.ServerUsername` instead of `Properties.ServerPassword`, use the same provider and parameter pattern with the `Properties.ServerUsername.*` prefix.

### 1Password-CLI

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.Item,Properties.ServerPassword.Parameters.Field,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.Item,Password.Parameters.Field,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,101,linux-service-account,password,,101,rfpkcs12-store,password,
```

### Azure-KeyVault

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretId,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.SecretId,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,102,linux-service-account-password,,102,rfpkcs12-store-password,
```

### Azure-KeyVault-ServicePrincipal

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretId,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.SecretId,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,103,linux-service-account-password,,103,rfpkcs12-store-password,
```

### BeyondTrust-PasswordSafe

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SystemId,Properties.ServerPassword.Parameters.AccountId,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.SystemId,Password.Parameters.AccountId,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,104,bt-system-123,bt-account-456,,104,bt-system-123,bt-account-789,
```

### CyberArk-CentralCredentialProvider

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.Safe,Properties.ServerPassword.Parameters.Folder,Properties.ServerPassword.Parameters.Object,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.Safe,Password.Parameters.Folder,Password.Parameters.Object,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,105,Certificates,Root,linux-service-account,,105,Certificates,Root,rfpkcs12-store-password,
```

### CyberArk-SdkCredentialProvider

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.Safe,Properties.ServerPassword.Parameters.Folder,Properties.ServerPassword.Parameters.Object,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.Safe,Password.Parameters.Folder,Password.Parameters.Object,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,106,Certificates,Root,linux-service-account,,106,Certificates,Root,rfpkcs12-store-password,
```

### Delinea-SecretServer

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretId,Properties.ServerPassword.Parameters.SecretFieldName,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.SecretId,Password.Parameters.SecretFieldName,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,107,12001,password,,107,12002,password,
```

### GCP-SecretManager

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.secretId,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.secretId,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,108,linux-service-account-password,,108,rfpkcs12-store-password,
```

### Hashicorp-Vault

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.Secret,Properties.ServerPassword.Parameters.Key,Properties.ServerPassword.SecretValue,Password.ProviderId,Password.Parameters.Secret,Password.Parameters.Key,Password.SecretValue
00000000-0000-0000-0000-000000000001,linux01.example.com,/opt/certs/app.p12,109,certstores/linux01,serverPassword,,109,certstores/linux01,storePassword,
```

## Step 4: Sync The Migration

Run the import command with `--sync`:

```bash
kfutil stores import csv \
  --file K8SCluster_pam_sync.csv \
  --store-type-name K8SCluster \
  --sync \
  --no-prompt
```

Use one command per store type CSV.

## Step 5: Verify The Migration

Export the store type again:

```bash
kfutil stores export --store-type-name K8SCluster
```

Confirm the migrated rows include:

```text
Properties.ServerPassword.Provider
Properties.ServerPassword.Parameters.<ParameterName>
```

Confirm `Properties.ServerPassword.SecretValue` is empty or absent for migrated rows.

Review the sync results file and confirm the `Errors` column is empty for each migrated row.

## Notes

- This workflow changes where Keyfactor retrieves the store credential. It does not rotate the credential in the target system.
- When moving the other direction, from PAM-backed credentials to static credentials, put JSON secrets in one CSV cell and escape inner quotes by doubling them, for example `"{""kind"":""Config""}"`.
- Non-JSON static secrets can be written directly in the credential column, with normal CSV quoting when the value contains commas, quotes, or line breaks.
- For provider-backed `ServerUsername`, use the same pattern with `Properties.ServerUsername.Provider` and `Properties.ServerUsername.Parameters.*`.
- For store-level passwords, use `Password.ProviderId` and `Password.Parameters.*`.
- Test with one store before applying the same provider values to many stores.

## Related Commands

- [kfutil stores export](../kfutil_stores_export.md)
- [kfutil stores import csv](../kfutil_stores_import_csv.md)
- [Bulk Certificate Store Updates](bulk-certificate-store-updates.md)
