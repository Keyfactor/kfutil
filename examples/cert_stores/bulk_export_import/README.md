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
```bash
kfutil store-types create                                          
? Choose an option:  [Use arrows to move, type to filter]
> AKV
  AzureApp
  AzureAppGW
  AzureSP
  Fortigate
  HCVKV
  HCVKVJKS
```

#### Non-Interactive Mode:
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
```bash
kfutil stores export --store-type-name k8ssecret
```

### Step 2: Edit the CSV file
This can be hard without an example, so it's recommended that you create a store in Keyfactor Command and then export 
it to see what the CSV file should look like. At bare minimum, the store credentials will need to be filled in as they
cannot be exported from Keyfactor Command.

### Step 3: Import the CSV file
You can use `kfutil` to import a CSV file that contains the certificate store information to be imported into Keyfactor Command.
```bash
kfutil stores import --file /path/to/csv/file.csv
```