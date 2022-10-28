## Quickstart

```bash
make install
kfutil --help
````

### Environmental Variables

All the variables listed below need to be set in your environment. The `kfutil` command will look for these variables
and use them if they are set. If they are not set, the utility will fail to connect to Keyfactor.

```bash
export KEYFACTOR_HOSTNAME=<mykeyfactorhost.mydomain.com>
export KEYFACTOR_USERNAME=<myusername> # Do not include domain
export KEYFACTOR_PASSWORD=<mypassword>
export KEYFACTOR_DOMAIN=<mykeyfactordomain>
```

## Commands

### Bulk operations

#### Bulk create cert stores

`# TODO: Not implemented`  
This will attempt to process a CSV input file of certificate stores to create. The template can be generated by
running: `kfutil generate-template --type bulk-certstore` command.

```bash
kfutil bulk create certstores --file <path to csv file>
```

#### Bulk create cert store types

`# TODO: Not implemented`
This will attempt to process a CSV input file of certificate store types to create. The template can be generated by
running: `kfutil generate-template --type bulk-certstore-types` command.

```bash
kfutil bulk create certstores --file <path to csv file>
```

### Root of Trust
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

This will write the file `certs_template.csv` to the current directory.

```bash
kfutil stores generate-template-rot --type certs
```

#### Generate Certificate Store List Template

This will write the file `stores_template.csv` to the current directory.

```bash
kfutil stores generate-template-rot --type stores
```

#### Run Root of Trust Audit

Audit will take in a list of certificates and a list of certificate stores and check that the certificate store's 
inventory either contains the certificate or does not contain the certificate based on the `--add-certs` and 
`--remove-certs` flags. These flags can be used together or separately. The aforementioned flags take in a path to CSV 
files containing a list of certificate thumbprints. To generate a template for these files, run the following command:
```bash
kfutil stores rot generate-template --type certs
```
In addition, you must provide a list of stores you wish to audit. To generate a template for this file, run the following
command:
```bash
kfutil stores rot generate-template --type stores
```
With all the files generated and populated, you can now run the audit command:
```bash
kfutil stores rot audit \
  --stores stores.csv \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv
```
This will generate an audit file that contains the results of the audit and actions that will be taken if `reconcile` is
executed. By default, the audit file will be named `rot_audit.csv` and will be written to the current directory. To output
the audit file to a different location, use the `--output` flag:
```bash
kfutil stores rot audit \
  --stores stores.csv \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --output /path/to/output/autdit_file.csv
```


#### Run Root of Trust Reconcile

Reconcile will take in a list of certificates and a list of certificate stores and check that the certificate store's
inventory either contains the certificate or does not contain the certificate based on the `--add-certs` and
`--remove-certs` flags. These flags can be used together or separately. The aforementioned flags take in a path to CSV
files containing a list of certificate thumbprints. To generate a template for these files, run the following command:
```bash
kfutil stores rot generate-template --type certs
```
In addition, you must provide a list of stores you wish to reconcile. To generate a template for this file, run the following
command:
```bash
kfutil stores rot generate-template --type stores
```
With all the files generated and populated, you can now run the reconcile command:
```bash
kfutil stores rot reconcile \
  --stores stores.csv \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv
```
This will generate an audit file that contains the results of the audit and actions will immediately execute those actions.
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

### Development

This CLI developed using [cobra](https://umarcor.github.io/cobra/)

#### Adding a new command

```bash
cobra-cli add <my-new-command>
```

alternatively you can specify the parent command

```bash
cobra-cli add <my-new-command> -p '<parent>Cmd'
```