#!/usr/bin/env bash
set -e

kfutil logout || true

export DEMO_CERT_CN="CommandCA1" # Change this to the CN of the cert you want to add to all stores

#echo "Sourcing Keyfactor command environment variables from ~/.env_kf-int-lab1022..."
source ~/.env_kf-int-lab1022

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

echo "Re-auditing stores to confirm reconciliation..."
kfutil stores rot audit \
  --add-certs certs_template.csv \
  --stores stores_template.csv

echo "Opening the audit report for you to review/make changes..."
read -p "Press enter to continue"
open rot_audit.csv
read -p "Press enter to continue"

echo "Resetting demo configuration to original state..."

echo "Opening the reset audit report for you to review/make changes..."
read -p "Press enter to continue"
open rot_audit_reset.csv

kfutil stores rot reconcile \
  --import-csv \
  --input-file rot_audit_reset.csv

echo "Reset complete. Re-run the demo script to re-audit and re-apply the demo configuration."