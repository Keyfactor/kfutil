# Keyfactor Command Bulk Certificate Store Operations
This will demo how to use `kfutil` to bulk import and export Keyfactor Command Certificate Stores.

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Guide](#guide)
  * [Step 0: Create the certificate store type](#step-0-create-the-certificate-store-type)
    + [User Interactive Mode:](#user-interactive-mode)
      - [Example Output:](#example-output)
    + [Non-Interactive Mode:](#non-interactive-mode)
  * [Step 1: Generate an import template CSV file](#step-1-generate-an-import-template-csv-file)
    + [Export a store-type template](#export-a-store-type-template)
    + [Export an existing certificate stores by store-type](#export-an-existing-certificate-stores-by-store-type)
      - [User Interactive Mode:](#user-interactive-mode-1)
        * [Example Output:](#example-output-1)
      - [Non-Interactive Mode:](#non-interactive-mode-1)
        * [Example Output:](#example-output-2)
  * [Step 2: Edit the CSV file](#step-2-edit-the-csv-file)
  * [Step 3: Import the CSV file](#step-3-import-the-csv-file)
    + [Example Output:](#example-output-3)
- [Operations](#operations)
  * [Bulk Create/Import Certificate Stores](#bulk-createimport-certificate-stores)
    + [User Interactive Mode:](#user-interactive-mode-2)
    + [Non-Interactive Mode:](#non-interactive-mode-2)
    + [Example Output:](#example-output-4)
  * [Bulk Export Certificate Stores](#bulk-export-certificate-stores)
    + [User Interactive Mode:](#user-interactive-mode-3)
    + [Non-Interactive Mode:](#non-interactive-mode-3)
    + [Example Output:](#example-output-5)
  * [Delete Certificate Store](#delete-certificate-store)
    + [User Interactive Mode:](#user-interactive-mode-4)
    + [Non-Interactive Mode:](#non-interactive-mode-4)
    + [Example Output:](#example-output-6)

<!-- tocstop -->

## Prerequisites
- `kfutil` v1.4.0 or later
- Keyfactor Command v10.x
- The certificate store type exists in Keyfactor Command

## Guide

### Step 0: Create the certificate store type
You can use `kfutil` to create a specific certificate store type, or all supported certificate store types. 

#### User Interactive Mode:
Below is an example of creating a specific certificate store type `AKV`.
```bash
kfutil store-types create
```
##### Example Output:
```text
kfutil store-types create                                          
? Choose an option:  [Use arrows to move, type to filter]
> AKV
  AzureApp
  AzureAppGW
  AzureSP
  Fortigate
  HCVKV
  HCVKVJKS
  
Certificate store type AKV created with ID: 166
```

#### Non-Interactive Mode:
This will create all supported certificate store types. *Note* this will not update any existing certificate store types.
```bash
kfutil store-types create --all
```

### Step 1: Generate an import template CSV file
You can use `kfutil` to generate a CSV file that can be used to bulk import certificate stores into Keyfactor Command.
*NOTE*: The certificate store and/or type must exist in Keyfactor Command before you can generate the CSV file.

#### Export a store-type template
[!NOTE] This will not export any values, just the headers associated with the store type. It is highly recommended to 
create a store in Keyfactor Command and then export it to see what the values of the CSV file look like.
```bash
kfutil stores import generate-template --store-type-name rfpem
```
#### Export an existing certificate stores by store-type
Below is an example of exporting all certificate stores of a specific store type `k8ssecret`.
[!IMPORTANT] This will *not* export any secrets or sensitive information associated with the certificate stores.

##### User Interactive Mode:
```bash
kfutil stores export      
```

###### Example Output:
```text
kfutil stores export
? Choose a store type to export:  [Use arrows to move, type to filter]
> K8SSecret
  K8SCluster
  K8SNS

Stores exported for store type with id 133 written to K8SSecret_stores_export_1709065204.csv
```

##### Non-Interactive Mode:
This will export all certificate stores of all store types into CSV files for each store type.
```bash
kfutil stores export --all
```

###### Example Output:
```text
kfutil stores export --all

Stores exported for store type with id 133 written to K8SSecret_stores_export_1709066753.csv

Stores exported for store type with id 147 written to K8SCluster_stores_export_1709066754.csv

Stores exported for store type with id 149 written to K8SNS_stores_export_1709066755.csv
```

### Step 2: Edit the CSV file
This can be hard without an example, so it's recommended that you create a store in Keyfactor Command and then export 
it to see what the CSV file should look like. At bare minimum, the store credentials will need to be filled in as they
cannot be exported from Keyfactor Command.

### Step 3: Import the CSV file
You can use `kfutil` to import a CSV file that contains the certificate store information to be imported into Keyfactor 
Command.
```bash
kfutil stores import csv --file /path/to/csv/file.csv
```

#### Example Output:
```text
 kfutil stores import csv --file K8SCluster_stores_export_1709066656.csv 
? Choose a store type to import:  [Use arrows to move, type to filter]
  K8SSecret
> K8SCluster
  K8SNS

1 records processed.
1 rows had errors.
Import results written to K8SCluster_stores_export_1709066656_results.csv
```

## Operations

### Bulk Create/Import Certificate Stores

#### User Interactive Mode:
```bash
kfutil stores import csv --file /path/to/csv/file.csv
```

#### Non-Interactive Mode: 
```bash
kfutil stores import csv \
  --file /path/to/csv/file.csv \
  --store-type-name <store-type-short-name>
```

#### Example Output:
```text
 kfutil stores import csv --file K8SCluster_stores_export_1709066656.csv
? Choose a store type to import:  [Use arrows to move, type to filter]
> K8SCluster
? Choose a store type to import: K8SCluster
1 records processed.
1 rows had errors.
Import results written to K8SCluster_stores_export_1709072563_results.csv
```

### Bulk Export Certificate Stores

#### User Interactive Mode:
```bash
kfutil stores export
```

#### Non-Interactive Mode:
```bash
kfutil stores export --all
```

#### Example Output:
```text
kfutil stores export      
? Choose a store type to export:  [Use arrows to move, type to filter]
> K8SCluster
? Choose a store type to export: K8SCluster

Stores exported for store type with id 147 written to K8SCluster_stores_export_1709072563.csv
```

### Delete Certificate Store

#### User Interactive Mode:
```bash
kfutil stores delete
```

#### Non-Interactive Mode:
```bash
kfutil stores delete --file /path/to/csv/file.csv
```

#### Example Output:
```text
kfutil stores delete
? Choose a store type to import:  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
> [x]  K8SSecret/dev/keyfactor/test (d2de98e9-b384-4b55-9118-d5780655aeee)
> [ ]  K8SCluster/dev/cluster (a78b08e1-063b-42da-8c97-81a0bf8cd282)
> [x]  K8SNS/dev/default (75b732b9-71a1-4ff4-ac70-60a8c5936532)
Choose stores to delete: K8SSecret/dev/keyfactor/test (d2de98e9-b384-4b55-9118-d5780655aeee), K8SNS/dev/default (75b732b9-71a1-4ff4-ac70-60a8c5936532)
successfully deleted store d2de98e9-b384-4b55-9118-d5780655aeee
successfully deleted store 75b732b9-71a1-4ff4-ac70-60a8c5936532
```