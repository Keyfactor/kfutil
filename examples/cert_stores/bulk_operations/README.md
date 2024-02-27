# Keyfactor Command Bulk Certificate Store Operations
This will demo how to use `kfutil` to bulk import and export Keyfactor Command Certificate Stores.

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

