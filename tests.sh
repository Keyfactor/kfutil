#!/usr/bin/env bash

set -e
# Root of Trust
## Generate a certs template file prepopulated with certs where CN=terraform or CN=terraformer
echo "Generating certs template file"
kfutil stores rot generate-template \
  --type certs \
  --cn terraform \
  --cn terraformer

echo "Show file: certs_template.csv"
cat certs_template.csv

## Same as above but writing to non-default file
echo "Generating certs template file and writing to $(date +"%Y_%m_%d")_certs.csv"
kfutil stores rot generate-template \
  --type certs \
  --cn terraform \
  --cn terraformer \
  --outpath $(date +"%Y_%m_%d")_certs.csv

echo "Show file: $(date +"%Y_%m_%d")_certs.csv"
cat $(date +"%Y_%m_%d")_certs.csv

## Generate a stores template file prepopulated with all stores
echo "Generating stores template file"
kfutil stores rot generate-template \
  --type stores \
  --store-type all

echo "Show file: stores_template.csv"
cat stores_template.csv

## Same as above but writing to non-default file
echo "Generating stores template file and writing to $(date +"%Y_%m_%d")_stores.csv"
kfutil stores rot generate-template \
  --type stores \
  --store-type all \
  --outpath $(date +"%Y_%m_%d")_stores.csv

echo "Show file: $(date +"%Y_%m_%d")_stores.csv"
cat $(date +"%Y_%m_%d")_stores.csv
## Generate audit report for scenario where terraform and terraformer are compromised and you want to remove them from all stores
echo "Generating audit report for scenario where terraform and terraformer are compromised and you want to remove them from all stores"
kfutil stores rot audit \
  --remove-certs certs_template.csv \
  --stores stores_template.csv

echo "Show file: rot_audit.csv"
cat rot_audit.csv

## Same as above but saved to non-default file
echo "Generating audit report for scenario where terraform and terraformer are compromised and you want to remove them from all stores and writing to $(date +"%Y_%m_%d")_audit_report.csv"
kfutil stores rot audit \
  --remove-certs certs_template.csv \
  --stores stores_template.csv \
  --outpath $(date +"%Y_%m_%d")_audit_report.csv

echo "Show file: $(date +"%Y_%m_%d")_audit_report.csv"
cat $(date +"%Y_%m_%d")_audit_report.csv

## Reconcile the report via import
echo "Reconciling the report via import"
kfutil stores rot reconcile --import-csv

## Same as above but test non-default file
echo "Reconciling the report via import and reading from $(date +"%Y_%m_%d")_audit_report.csv"
kfutil stores rot reconcile --import-csv --input-file $(date +"%Y_%m_%d")_audit_report.csv

## Alternatively an ad-hoc reconcile. This generates the report based on the input files and then reconciles the report.
echo "Reconciling the report via ad-hoc reconcile"
kfutil stores rot reconcile \
  --remove-certs certs_template.csv \
  --stores stores_template.csv

echo "Show file: rot_audit.csv"
cat rot_audit.csv

## Same as above but test non-default file
echo "Reconciling the report via ad-hoc reconcile and reading from $(date +"%Y_%m_%d")_certs.csv and $(date +"%Y_%m_%d")_stores.csv"
kfutil stores rot reconcile \
  --remove-certs $(date +"%Y_%m_%d")_certs.csv \
  --stores $(date +"%Y_%m_%d")_stores.csv \
  --outpath $(date +"%Y_%m_%d")_audit_report.csv

echo "Show file: $(date +"%Y_%m_%d")_audit_report.csv"
cat $(date +"%Y_%m_%d")_audit_report.csv

## Add the certs back for next test
echo "Adding the certs back"
kfutil stores rot reconcile \
  --add-certs certs_template.csv \
  --stores stores_template.csv

## Same as above but test non-default file
echo "Adding the certs back and reading from $(date +"%Y_%m_%d")_certs.csv and $(date +"%Y_%m_%d")_stores.csv"
kfutil stores rot reconcile \
  --add-certs $(date +"%Y_%m_%d")_certs.csv \
  --stores $(date +"%Y_%m_%d")_stores.csv

## Clean all generated file up
echo "Cleaning up"
rm -f $(date +"%Y_%m_%d")_certs.csv \
  $(date +"%Y_%m_%d")_stores.csv \
  gen_audit_report.csv \
  certs_template.csv \
  stores_template.csv \
  rot_audit.csv