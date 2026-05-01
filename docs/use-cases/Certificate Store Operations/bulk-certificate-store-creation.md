# Bulk Certificate Store Creation

Use this workflow when you need to create many certificate stores of the same type from a CSV file.

This example creates ten Kubernetes certificate stores:

- Five `K8SSecret` stores.
- Five `K8STLSSecr` stores.
- Three stores of each type use static Keyfactor-encrypted credentials.
- Two stores of each type use a PAM provider-backed `ServerPassword`.

## Contents

- [Before You Begin](#before-you-begin)
- [Step 1: Choose The Store Types](#step-1-choose-the-store-types)
- [Step 2: Prepare Static Credential Rows](#step-2-prepare-static-credential-rows)
- [Step 3: Prepare PAM Provider Rows](#step-3-prepare-pam-provider-rows)
- [Step 4: Create K8SSecret Stores](#step-4-create-k8ssecret-stores)
- [Step 5: Create K8STLSSecr Stores](#step-5-create-k8stlssecr-stores)
- [Step 6: Verify The Created Stores](#step-6-verify-the-created-stores)
- [Notes](#notes)
- [Related Commands](#related-commands)

## Before You Begin

You need:

- `kfutil` configured to authenticate to Keyfactor Command.
- Permission to create certificate stores.
- The target certificate store types already created in Command.
- A registered orchestrator agent ID.
- Static credential values or a configured PAM provider.

For Kubernetes stores, `ClientMachine` should match the orchestrator target expected by the extension, and `StorePath` should identify the Kubernetes namespace and secret name.

## Step 1: Choose The Store Types

This demo uses:

```text
K8SSecret
K8STLSSecr
```

Each type gets its own CSV because `kfutil stores import csv` accepts one store type per command.

## Step 2: Prepare Static Credential Rows

Static credential rows use direct credential columns:

```text
Properties.ServerUsername
Properties.ServerPassword
```

Example `K8SSecret` static row:

```csv
ContainerId,ClientMachine,StorePath,CreateIfMissing,Properties.KubeSecretName,Properties.KubeSecretType,Properties.IncludeCertChain,Properties.SeparateChain,Properties.ServerUseSsl,AgentId,Properties.ServerUsername,Properties.ServerPassword
0,kf-integrations,default/kfutil-demo-k8ssecret-1,true,kfutil-demo-k8ssecret-1,secret,true,true,true,275bcd31-9e7b-4c4a-bce9-1719e0c2168d,kubeconfig,"<kubeconfig-json>"
```

If the credential value is JSON, keep it as a CSV string. `kfutil` treats credential fields as secret strings even when the cell value looks like JSON.

## Step 3: Prepare PAM Provider Rows

PAM-backed rows use provider columns instead of a direct `Properties.ServerPassword` value:

```text
Properties.ServerPassword.Provider
Properties.ServerPassword.Parameters.SecretName
Properties.ServerPassword.Parameters.SecretType
Properties.ServerPassword.Parameters.StaticSecretFieldName
```

Example `K8SSecret` PAM row:

```csv
ContainerId,ClientMachine,StorePath,CreateIfMissing,Properties.KubeSecretName,Properties.KubeSecretType,Properties.IncludeCertChain,Properties.SeparateChain,Properties.ServerUseSsl,AgentId,Properties.ServerUsername,Properties.ServerPassword.Provider,Properties.ServerPassword.Parameters.SecretName,Properties.ServerPassword.Parameters.SecretType,Properties.ServerPassword.Parameters.StaticSecretFieldName
0,kf-integrations,default/kfutil-demo-k8ssecret-4,true,kfutil-demo-k8ssecret-4,secret,true,true,true,275bcd31-9e7b-4c4a-bce9-1719e0c2168d,kubeconfig,30,dev/aks/kf-integrations,static_json," "
```

The provider ID and parameter names depend on your PAM provider type.

## Step 4: Create K8SSecret Stores

Create a CSV named `k8ssecret_bulk_create.csv` with five rows:

- Rows 1-3 use `Properties.ServerPassword`.
- Rows 4-5 use `Properties.ServerPassword.Provider` and `Properties.ServerPassword.Parameters.*`.

Run:

```bash
kfutil stores import csv \
  --file k8ssecret_bulk_create.csv \
  --store-type-name K8SSecret \
  --no-prompt \
  --results-path k8ssecret_bulk_create_results.csv
```

Expected output:

```text
5 records processed.
5 certificate stores successfully created.
Import results written to k8ssecret_bulk_create_results.csv
```

## Step 5: Create K8STLSSecr Stores

Create a CSV named `k8stlssecr_bulk_create.csv` with five rows. Use the same credential pattern, but set the Kubernetes secret type values for TLS secret stores.

Run:

```bash
kfutil stores import csv \
  --file k8stlssecr_bulk_create.csv \
  --store-type-name K8STLSSecr \
  --no-prompt \
  --results-path k8stlssecr_bulk_create_results.csv
```

Expected output:

```text
5 records processed.
5 certificate stores successfully created.
Import results written to k8stlssecr_bulk_create_results.csv
```

## Step 6: Verify The Created Stores

Export each store type:

```bash
kfutil stores export --store-type-name K8SSecret
kfutil stores export --store-type-name K8STLSSecr
```

Verify that the five new rows for each store type are present.

For the static rows, confirm that `Properties.ServerPassword.SecretValue` is present in the export.

For the PAM-backed rows, confirm that `Properties.ServerPassword.Provider` and the expected `Properties.ServerPassword.Parameters.*` columns are present.

## Notes

- Use unique `StorePath` and `Properties.KubeSecretName` values for each row.
- Keep one CSV per store type.
- Check the `Errors` column in the results CSV after every import.
- CSV files may contain sensitive credentials. Protect the input and results files according to your operating procedures.

## Related Commands

- [kfutil stores import csv](../../kfutil_stores_import_csv.md)
- [kfutil stores import generate-template](../../kfutil_stores_import_generate-template.md)
- [kfutil stores export](../../kfutil_stores_export.md)
- [Bulk Certificate Store Updates](bulk-certificate-store-updates.md)
