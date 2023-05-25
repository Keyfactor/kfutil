# Root of Trust Management

## Overview
Q. What is a "root of trust" or "trusted root"?
A. A root of trust is a trusted entity from which a chain of trust is derived. The root of trust is the foundation of 
the chain of trust, and in the realm of PKI, the root of a trust is a root certificate authority (CA).

Q. What is a "chain of trust"?
A. A chain of trust is a sequence of entities that vouch for the identity of the next entity in the chain. In the realm
of PKI, a chain of trust is a sequence of certificates, where each certificate is signed by the entity identified by the
next certificate in the chain.

Q. How can Keyfactor Command be used to manage the root(s) of trust?
A. This is what this example below is all about. The example will show how to use the certificate stores you've 
already configured in Keyfactor Command to audit and manage the root(s) of trust in your environment.

## Prerequisites
* Keyfactor Command 10.x or later
* [kfutil](https://github.com/Keyfactor/kfutil/releases) `v1.0.0` or later
* [jq](https://stedolan.github.io/jq/) `v1.6` or later
* Keyfactor Command API credentials with appropriate permissions:
  * read/write access to Certificate Stores
  * read access to Certificates
  * read access to Containers
  * read access to CertificateStoreTypes
  * read access to Agents
* Keyfactor Command Orchestrator configured and running.
* Keyfactor Command Container(s) configured if you're using them.
* Keyfactor Command Certificate Store(s) created.
* Keyfactor Command Certificate(s) created.

## Demo
This demo will show how to use `kfutil` to audit and manage certificate store inventory in your Keyfactor Command 
environment. The demo will use `kfutil stores rot` commands to audit and manage stores represented as "roots of trust" (`rot`), 
using Keyfactor Command Containers. Containers are a way to organize certificate stores in Keyfactor Command. **It is 
assumed that the `orchestrators`, `stores`, `certificates`, and `containers` have already been created and configured in 
Keyfactor Command.**
```bash
#!/usr/bin/env bash
set -e

kfutil logout || true

export DEMO_CERT_CN="CommandCA1" # Change this to the CN of the cert you want to add to all stores

#echo "Sourcing Keyfactor command environment variables from ~/.env_kf-int-lab1022..."
#source ~/.env_kf-int-lab1022

# If KEYFACTOR_HOSTNAME is not set, prompt for it
echo "Setting up Keyfactor command environment variables..."
if [[ -z "$KEYFACTOR_API_PATH" ]]; then
 export KEYFACTOR_API_PATH="KeyfactorAPI"
fi

if [[ -z "$KEYFACTOR_HOSTNAME" || -z "$KEYFACTOR_USERNAME" || -z "$KEYFACTOR_USERNAME" || -z "$KEYFACTOR_PASSWORD" || -z "$KEYFACTOR_DOMAIN" ]]; then
  echo "One or more Keyfactor command environment variables are not set, will use 'kfutil login' to prompt for them..."
  kfutil login
else
  echo "Using existing Keyfactor command environment variables..."
  kfutil login --no-prompt
fi
read -p "Press enter to continue"

# If you'd rather be prompted for everything above this line, use this instead:
#kfutil login

#echo "List of stores from keyfactor API as json..."
#kfutil stores list --exp
#read -p "Press enter to continue"

#echo "Add my Keyfactor Command root CA to all my stores"
#echo "Populate stores CSV with all stores..."
#kfutil stores rot generate-template \
#  --type stores \
#  --store-type all

#echo "Populate stores CSV with stores of type RFPEM or K8SPKCS12..."
#kfutil stores rot generate-template \
#  --type stores \
#  --store-type RFPEM \
#  --store-type K8SPKCS12

echo "Populate stores CSV where the container names are 'Trusted Roots', 'Intermediate CAs', or 'RootStores'..."
kfutil stores rot generate-template \
  --type stores \
  --container-name "Trusted Roots" \
  --container-name "Intermediate CAs" \
  --container-name "RootStores"

echo "Opening the stores template..."
read -p "Press enter to continue"
open stores_template.csv
read -p "Press enter to continue"

echo "Populate certs.csv with my Keyfactor Command root CA 'CN=$DEMO_CERT_CN'..."
kfutil stores rot generate-template \
  --type certs \
  --cn "$DEMO_CERT_CN"

echo "Opening the certs template for you to review/make changes..."
read -p "Press enter to continue"
open certs_template.csv
read -p "Press enter to continue"

echo "Run audit/dry run of what 'reconcile' will do and writing it to report file..."
kfutil stores rot audit \
  --add-certs certs_template.csv \
  --stores stores_template.csv

echo "Opening the audit report for you to review/make changes..."
read -p "Press enter to continue"
open rot_audit.csv
read -p "Press enter to continue"

echo "Reconcile stores using the audit report and any changes you made to it..."
kfutil stores rot reconcile \
  --import-csv rot_audit.csv
```

## Outline
1. Use `kfutil` to generate a CSV template of certificate stores you want to audit. There is an option to specify all stores,
   (`--store-type all`) or you can specify a list of stores by store type name, or you can specify a list of stores by 
container name. The container name is the name of the container in Keyfactor Command that the store is in. 
2. Use `kfutil` to generate a CSV template and pre-populated with certificates you want to audit. You can specify a list of 
certificates by common name (CN) or thumbprint. 
3. Use `kfutil` to audit the stores and certificates. This will generate a CSV report of what changes will be made to the
certificate stores.
4. Use `kfutil` to reconcile the stores with the audit actions. This will schedule Keyfactor Command Orchestrator jobs 
to be scheduled to make the changes to the certificate stores.
5. Wait for the Orchestrator jobs to complete. You can view the jobs in the Keyfactor Command Portal, `Orchestrators > Jobs` 
section of the UI.
6. Use `kfutil` to audit the stores and certificates again to verify the changes were made as expected. The report will
read as no changes are needed (`add` and `remove` cols are both entirely `false`).
7. Repeat steps 1-6 as needed to manage the root(s) of trust, or any other types of certificate stores, as this utility
can be used as a general purpose, bulk, certificate store inventory management tool.


