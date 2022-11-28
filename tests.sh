#!/usr/bin/env bash
# Root of Trust
set -e

TEST_CERT_CN_1="failing.personality.org"
TEST_CERT_CN_2="jar.royalty.io"

function test_generate_cert_templates() {
    ## Generate a certs template file prepopulated with certs where CN=$TEST_CERT_CN_1 or CN=$TEST_CERT_CN_2
    echo "Generating certs template file"
    kfutil stores rot generate-template \
      --type certs \
      --cn $TEST_CERT_CN_1 \
      --cn $TEST_CERT_CN_2

    echo "Show file: certs_template.csv"
    cat certs_template.csv

    ## Same as above but writing to non-default file
    echo "Generating certs template file and writing to $(date +"%Y_%m_%d")_certs.csv"
    kfutil stores rot generate-template \
      --type certs \
      --cn $TEST_CERT_CN_1 \
      --cn $TEST_CERT_CN_2 \
      --outpath $(date +"%Y_%m_%d")_certs.csv

    echo "Show file: $(date +"%Y_%m_%d")_certs.csv"
    cat $(date +"%Y_%m_%d")_certs.csv
}

function test_generate_store_templates() {
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
}

function test_generate_audit_reports(){
  ## Generate audit report for scenario where $TEST_CERT_CN_1 and $TEST_CERT_CN_2 are compromised and you want to remove them from all stores
  echo "Generating audit report for scenario where $TEST_CERT_CN_1 and $TEST_CERT_CN_2 are compromised and you want to remove them from all stores"
  kfutil stores rot audit \
    --remove-certs certs_template.csv \
    --stores stores_template.csv

  echo "Show file: rot_audit.csv"
  cat rot_audit.csv

  ## Same as above but saved to non-default file
  echo "Generating audit report for scenario where $TEST_CERT_CN_1 and $TEST_CERT_CN_2 are compromised and you want to remove them from all stores and writing to $(date +"%Y_%m_%d")_audit_report.csv"
  kfutil stores rot audit \
    --remove-certs certs_template.csv \
    --stores stores_template.csv \
    --outpath $(date +"%Y_%m_%d")_audit_report.csv

  echo "Show file: $(date +"%Y_%m_%d")_audit_report.csv"
  cat $(date +"%Y_%m_%d")_audit_report.csv
}

function test_reconcile(){
  ## Reconcile the report via import
  echo "Reconciling the report via import"
  kfutil stores rot reconcile --import-csv
  test_sleep


  ## Same as above but test non-default file
  echo "Reconciling the report via import and reading from $(date +"%Y_%m_%d")_audit_report.csv"
  kfutil stores rot reconcile --import-csv --input-file $(date +"%Y_%m_%d")_audit_report.csv
  test_sleep

  ## Alternatively an ad-hoc reconcile. This generates the report based on the input files and then reconciles the report.
  echo "Reconciling the report via ad-hoc reconcile"
  kfutil stores rot reconcile \
    --remove-certs certs_template.csv \
    --stores stores_template.csv

  echo "Show file: rot_audit.csv"
  cat rot_audit.csv

  test_sleep
  ## Same as above but test non-default file
  echo "Reconciling the report via ad-hoc reconcile and reading from $(date +"%Y_%m_%d")_certs.csv and $(date +"%Y_%m_%d")_stores.csv"
  kfutil stores rot reconcile \
    --remove-certs $(date +"%Y_%m_%d")_certs.csv \
    --stores $(date +"%Y_%m_%d")_stores.csv \
    --outpath $(date +"%Y_%m_%d")_audit_report.csv

  echo "Show file: $(date +"%Y_%m_%d")_audit_report.csv"
  cat $(date +"%Y_%m_%d")_audit_report.csv
}

function test_add_certs(){
  ## Add the certs back for next test
  echo "Adding the certs back"
  kfutil stores rot reconcile \
    --add-certs certs_template.csv \
    --stores stores_template.csv

  test_sleep

  ## Same as above but test non-default file
  echo "Adding the certs back and reading from $(date +"%Y_%m_%d")_certs.csv and $(date +"%Y_%m_%d")_stores.csv"
  kfutil stores rot reconcile \
    --add-certs $(date +"%Y_%m_%d")_certs.csv \
    --stores $(date +"%Y_%m_%d")_stores.csv
}

function test_cleanup(){
  ### Clean all generated file up
  echo "Cleaning up"
  rm -f $(date +"%Y_%m_%d")_certs.csv \
    $(date +"%Y_%m_%d")_stores.csv \
    gen_audit_report.csv \
    certs_template.csv \
    stores_template.csv \
    rot_audit.csv
}

function test_sleep(){
  echo "Sleeping for 2 minutes to allow orchestrator jobs to complete"
  sleep 120
}

function test_stores_list(){
  echo "Listing all stores"
  kfutil stores list
  echo "Formatting previous output as JSON"
  kfutil stores list | jq ".[] | {id: .Id, path: .Storepath}"

}

function test_stores_inventory(){
  echo "Running LIST operations"
  stID=$(kfutil stores list | jq -r ".[0].Id")
  echo "Getting inventory for store $sID"
  kfutil stores inventory show --sid "$stID" | jq -r '.[].Inventory'
  cName=$(kfutil stores list | jq -r ".[0].ContainerName")
  echo "Getting inventory for stores of container name $cName"
  kfutil stores inventory show --container "$cName" | jq -r '.[].Inventory'
  host=$(kfutil stores list | jq -r ".[0].ClientMachine")
  kfutil stores inventory show --client "$host" | jq -r '.[].Inventory'
  storeType=$(kfutil stores list | jq -r ".[0].CertStoreType")
  kfutil stores inventory show --store-type "$storeType" | jq -r '.[].Inventory'

  echo "Running CLEAR operations"
  kfutil stores inventory clear --sid "$stID" || true
  kfutil stores inventory clear --sid "$stID" --dry-run
  kfutil stores inventory clear --sid "$stID" --force --dry-run
  kfutil stores inventory clear --sid "$stID" --force

  echo "Running ADD operations"

  certSubj=$(kfutil certs list | jq -r ".[0].IssuedDN")
  kfutil stores inventory add --sid "$stID" --cn "$certSubj" || true
  kfutil stores inventory add --sid "$stID" --cn "$certSubj" --dry-run
  kfutil stores inventory add --sid "$stID" --cn "$certSubj" --dry-run --force
  kfutil stores inventory add --sid "$stID" --cn "$certSubj" --force

  echo "Running REMOVE operations"
  certSubj=$(kfutil certs list | jq -r ".[0].IssuedDN")
  kfutil stores inventory remove --sid "$stID" --cn "$certSubj" || true
  kfutil stores inventory remove --sid "$stID" --cn "$certSubj" --dry-run
  kfutil stores inventory remove --sid "$stID" --cn "$certSubj" --dry-run --force
  kfutil stores inventory remove --sid "$stID" --cn "$certSubj" --force

}
#kfutil stores rot generate-template --type certs --cn failing.personality.org --cn jar.royalty.io
test_stores_list
test_generate_cert_templates
test_generate_store_templates
test_generate_audit_reports
#test_reconcile
# Verify the certs are removed
#test_sleep
test_add_certs
test_stores_inventory
test_add_certs
# Verify the certs are added back
test_cleanup