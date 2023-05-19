#!/usr/bin/env bash
kfutil logout

export DEMO_CERT_CN="CommandCA1" # Change this to the CN of the cert you want to add to all stores

echo "Sourving Keyfactor command environment variables from ~/.env_kf-int-lab1022..."
source ~/.env_kf-int-lab1022

#echo "Setting up Keyfactor command environment variables..."
#export KEYFACTOR_HOSTNAME=your.command.com
#export KEYFACTOR_USERNAME=user@domain
#export KEYFACTOR_PASSWORD=thisismyuserpassword1234
#export KEYFACTOR_DOMAIN=domain
#export KEYFACTOR_API_PATH=KeyfactorAPI

kfutil login --no-prompt

# If you'd rather be prompted for everything above this line, use this instead:
#kfutil login

echo "List of stores from keyfactor API as json..."
kfutil stores list --exp | jq -r .
read -p "Press enter to continue"

echo "Add my Keyfactor Command root CA to all my stores..."

#echo "Populate stores CSV with all stores..."
#kfutil stores rot generate-template \
#  --type stores \
#  --store-type all

#echo "Populate stores CSV with stores of type RFPEM or K8SPKCS12..."
#kfutil stores rot generate-template \
#  --type stores \
#  --store-type RFPEM \
#  --store-type K8SPKCS12

echo "Populate stores of the following containers..."
kfutil stores rot generate-template \
  --type stores \
  --container-name "Trusted Roots" \
  --container-name "Intermediate CAs" \
  --container-name "RootStores"
read -p "Press enter to continue"

echo "Opening the stores template..."
open stores_template.csv
read -p "Press enter to continue"



echo "Populate certs.csv with my Keyfactor Command root CA..."
kfutil stores rot generate-template \
  --type certs \
  --cn "$DEMO_CERT_CN"
read -p "Press enter to continue"

echo "Opening the certs template for you to review/make changes..."
open certs_template.csv
read -p "Press enter to continue"

echo "Run audit/dry run of what 'reconcile' will do and writing it to report file..."
kfutil stores rot audit \
  --add-certs certs_template.csv \
  --stores stores_template.csv

echo "Opening the audit report for you to review/make changes..."
open rot_audit.csv
read -p "Press enter to continue"

echo "Reconcile stores using the audit report and any changes you made to it..."
kfutil stores rot reconcile \
  --import-csv rot_audit.csv
