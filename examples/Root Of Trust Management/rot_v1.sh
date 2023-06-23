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
  --import-csv \
  --input-file rot_audit.csv