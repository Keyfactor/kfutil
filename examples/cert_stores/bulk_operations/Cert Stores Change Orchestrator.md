# Changing the registered orchestrator agent for multiple Cert Stores

This example demonstrates how to change the registered orchestrator agent for multiple certificate stores in Keyfactor
Command using the `kfutil` CLI tool. This is particularly useful when you need to update the orchestrator agent for a
large number of stores efficiently.

## Assumptions

- You have `kfutil` installed and configured to connect to your Keyfactor Command instance.
- You know the IDs of the Orchestrator Agents you want to switch to.
- You have permissions to export and update certificate stores in Keyfactor Command.

## Step 1: Export Certificate Stores

First, export the certificate stores that you want to update. This will create a CSV file containing the details of the
stores.

```bash
kfutil stores export --all
```

This will export all certificate stores to multiple CSV files based on their store types. Example:

```shell
kfutil stores export --all

Stores exported for store type with id 183 written to AwsCerManA_stores_export_1765829171.csv

Stores exported for store type with id 178 written to K8SJKS_stores_export_1765829172.csv

Stores exported for store type with id 180 written to K8SPKCS12_stores_export_1765829173.csv
```

## Step 2: Modify the CSV File

Open the exported CSV files in a spreadsheet editor or text editor. Locate the `AgentId` column and update the values
to the new Orchestrator Agent ID that you want to assign to each store.

## Step 3: Import the Updated CSV File

After updating the CSV files with the new Orchestrator Agent IDs, you can import them back into Keyfactor Command using
the following command:

```bash
kfutil stores import csv --file /path/to/updated/csv/file.csv --sync --no-prompt
```

The `--sync` flag ensures that the import operation updates existing stores rather than creating duplicates. The
`--no-prompt` flag allows the operation to run without user interaction.

Example:

```shell
kfutil stores import csv --file K8SPKCS12_stores_export_1765743627.csv --store-type-name K8SPKCS12 -z --no-prompt
11 records processed.
9 certificate stores successfully created and/or updated.
2 rows had errors.
Import results written to K8SPKCS12_stores_export_1765743627_results.csv
```

## Step 4: Verify the Changes

After the import is complete, verify that the certificate stores have been updated with the new Orchestrator Agent IDs.
You can do this by exporting the stores again or checking directly in the Keyfactor Command interface.

# FAQ

## Q: Where can I find the Orchestrator Agent IDs?

A: You can find the Orchestrator Agent IDs in the Keyfactor Command interface under the Orchestrator Agents section, or
you can get a full list by using `kfutil orchs list`[docs](../../../docs/kfutil_orchs.md).