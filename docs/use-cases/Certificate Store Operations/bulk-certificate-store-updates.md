# Bulk Certificate Store Updates

Use this workflow when you need to update many existing Keyfactor Command certificate stores from a CSV file instead of editing each store in the Command UI.

Common examples include:

- Moving stores to a different orchestrator agent.
- Updating inventory schedules.
- Changing store metadata such as client machine, store path, container, or store-type properties.
- Correcting repeated configuration values after onboarding, migration, or environment changes.

`kfutil` performs bulk certificate store updates through the CSV import command with the `--sync` flag. The usual flow is:

```text
export stores -> edit CSV -> sync import -> review results -> verify changes
```

## Contents

- [Before You Begin](#before-you-begin)
- [Step 1: Export Stores](#step-1-export-stores)
- [Step 2: Edit The CSV](#step-2-edit-the-csv)
- [Step 3: Sync The Updates](#step-3-sync-the-updates)
- [Step 4: Review Results](#step-4-review-results)
- [Step 5: Verify Changes](#step-5-verify-changes)
- [Credentials](#credentials)
- [PAM Provider Credentials](#pam-provider-credentials)
- [Template Option](#template-option)
- [Operational Guidance](#operational-guidance)
- [Related Commands](#related-commands)

## Before You Begin

You need:

- `kfutil` configured to authenticate to Keyfactor Command.
- Permission to list, export, and update certificate stores.
- The certificate store type already created in Command.
- The store type short name, or the store type ID.

Keep one CSV file per certificate store type. The import command accepts one `--store-type-name` or `--store-type-id` per run, and store-type-specific properties differ by type.

## Step 1: Export Stores

Export the stores you want to update. For a single store type, use the store type short name:

```bash
kfutil stores export --store-type-name K8SSecret
```

Or use the store type ID:

```bash
kfutil stores export --store-type-id 154
```

To export all stores, grouped into separate CSV files by store type:

```bash
kfutil stores export --all
```

The export command writes files named like:

```text
K8SSecret_stores_export_1765743627.csv
```

The exported CSV includes an `Id` column. Preserve this column for every row you want to update.

## Step 2: Edit The CSV

Open the exported CSV and edit only the fields you intend to change.

For example, to move stores to another orchestrator, update the `AgentId` column:

```csv
Id,ClientMachine,StorePath,AgentId
6d1c7e86-0000-0000-0000-000000000000,k8s-worker-01,default/web-tls,275bcd31-0000-0000-0000-000000000000
```

For store-type properties, edit the `Properties.<PropertyName>` columns exported for that store type, such as:

```text
Properties.KubeNamespace
Properties.KubeSecretName
Properties.KubeSecretType
Properties.IncludeCertChain
```

For schedules, use one schedule shape per row:

```text
InventorySchedule.Immediate
InventorySchedule.Interval.Minutes
InventorySchedule.Daily.Time
InventorySchedule.Weekly.Days
InventorySchedule.Weekly.Time
```

Do not remove `Id` for update rows. When `--sync` is used, rows with an `Id` are updated. Rows without an `Id` are treated as create requests.

## Step 3: Sync The Updates

Run the CSV import command with `--sync`:

```bash
kfutil stores import csv \
  --file K8SSecret_stores_export_1765743627.csv \
  --store-type-name K8SSecret \
  --sync \
  --no-prompt
```

The equivalent command using a store type ID is:

```bash
kfutil stores import csv \
  --file K8SSecret_stores_export_1765743627.csv \
  --store-type-id 154 \
  --sync \
  --no-prompt
```

Use `--results-path` to choose where the results CSV is written:

```bash
kfutil stores import csv \
  --file K8SSecret_stores_export_1765743627.csv \
  --store-type-name K8SSecret \
  --sync \
  --no-prompt \
  --results-path K8SSecret_update_results.csv
```

## Step 4: Review Results

The command prints a summary:

```text
1 records processed.
1 certificate stores successfully updated.
Import results written to K8SSecret_update_results.csv
```

By default, the results file is named from the input file:

```text
<input-file>_results.csv
```

The results CSV contains the original row data and an `Errors` column. Successful rows have an empty `Errors` value. Failed rows include the API error message and should be corrected before rerunning.

Bulk sync is row-based, not all-or-nothing. One failed row does not mean every row failed.

## Step 5: Verify Changes

Verify the update by exporting the store type again:

```bash
kfutil stores export --store-type-name K8SSecret
```

Compare the updated columns against the original export and results file. For spot checks, fetch an individual store by ID:

```bash
kfutil stores get --id 6d1c7e86-0000-0000-0000-000000000000
```

## Credentials

Credential values can be supplied in the CSV, by flags, by environment variables, or by interactive prompts.

CSV columns:

```text
Properties.ServerUsername
Properties.ServerPassword
Password
```

Flags:

```bash
--server-username
--server-password
--store-password
```

Environment variables:

```text
KFUTIL_CSV_SERVER_USERNAME
KFUTIL_CSV_SERVER_PASSWORD
KFUTIL_CSV_STORE_PASSWORD
```

Values in the CSV take precedence over flags, environment variables, and prompts.

Avoid putting secrets in CSV files unless your operating procedures allow it. If you do use CSV-based secrets, protect the file, results file, and shell history accordingly.

## PAM Provider Credentials

Certificate store credentials can also reference a Keyfactor PAM provider instead of carrying a direct secret value. This is supported for CSV create and sync workflows when the CSV uses the provider columns exported by `kfutil`.

For an existing PAM-backed store, export the store type and use the exported credential columns as the pattern:

```bash
kfutil stores export --store-type-name K8SCluster
```

A PAM-backed `ServerPassword` uses columns like:

```text
Properties.ServerPassword.Provider
Properties.ServerPassword.Parameters.SecretName
Properties.ServerPassword.Parameters.SecretType
Properties.ServerPassword.Parameters.StaticSecretFieldName
```

Example:

```csv
Id,ClientMachine,StorePath,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretName,Properties.ServerPassword.Parameters.SecretType,Properties.ServerPassword.Parameters.StaticSecretFieldName
13b0b2c5-eb27-4885-91ec-fad35d0268df,kf-integrations,fresh,30,dev/aks/kf-integrations,static_json," "
```

To convert direct `ServerPassword` values to the same PAM provider, add the provider columns if they are not already present, fill in the provider ID and parameters, and leave any direct `Properties.ServerPassword.SecretValue` column empty.

Then run the normal sync command:

```bash
kfutil stores import csv \
  --file K8SCluster_pam_sync.csv \
  --store-type-name K8SCluster \
  --sync \
  --no-prompt
```

For a new store, leave the `Id` column empty or omit it and provide the same provider-backed credential columns:

```bash
kfutil stores import csv \
  --file K8SCluster_create_with_pam.csv \
  --store-type-name K8SCluster \
  --no-prompt
```

If the exported CSV contains masked direct credential values such as `********************`, prefer changing only the PAM-backed credential columns you intend to update. Do not copy masked values into new rows as real secrets.

## Template Option

If you need a blank CSV for a store type instead of exporting existing stores, generate a template:

```bash
kfutil stores import generate-template \
  --store-type-name K8SSecret \
  --outpath K8SSecret_bulk_import_template.csv
```

The template includes the common certificate store columns and the properties required for the selected store type. For updates, exporting existing stores is usually safer because it includes the `Id` values needed by `--sync`.

## Operational Guidance

- Start with a small CSV containing one or two stores.
- Keep the original export unchanged as a rollback/reference artifact.
- Preserve the `Id` column for update rows.
- Keep separate CSV files per store type.
- Edit only the columns needed for the change.
- Review the results CSV before rerunning failed rows.
- Rerun only corrected failed rows when possible.

## Related Commands

- [kfutil stores export](../kfutil_stores_export.md)
- [kfutil stores import csv](../kfutil_stores_import_csv.md)
- [kfutil stores import generate-template](../kfutil_stores_import_generate-template.md)
- [kfutil stores get](../kfutil_stores_get.md)
- [Migrate Static Store Credentials To A PAM Provider](migrate-static-store-credentials-to-pam.md)
