# Migrate Static Store Credentials To A PAM Provider

Use this workflow when existing certificate stores have static Keyfactor-encrypted credential values and you want those stores to reference a Keyfactor PAM provider instead.

This is a specialized bulk certificate store update. The workflow uses exported CSV files, edits the `Properties.ServerPassword` credential columns, then syncs the changes back to Keyfactor Command.

## Contents

- [Before You Begin](#before-you-begin)
- [Step 1: Export Stores](#step-1-export-stores)
- [Step 2: Identify The PAM Provider Columns](#step-2-identify-the-pam-provider-columns)
- [Step 3: Build The Sync CSV](#step-3-build-the-sync-csv)
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
