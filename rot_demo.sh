#!/usr/bin/env bash
# ROT Help
kfutil stores rot --help

# Create a report of actions the utility will take based on inputs from addCerts.csv and removeCerts.csv. Actions added
# to the report will also fit the criteria of a max of 1 private key, a minimum of 3 certs in the inventory and a
# maximum of 1 leaf cert.
kfutil stores reconcile audit \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --max-keys 1 \
  --min-certs 3 \
  --max-leaf-certs 1 \
  --stores stores.csv

# Add certs listed in addCerts.csv to stores listed if they meet the criteria of a max of 1 private key, a minimum of 3
# certs in the inventory and a maximum of 1 leaf cert. Then remove the certs listed in removeCerts.csv from the stores.
kfutil stores rot reconcile \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --max-keys 1 \
  --min-certs 3 \
  --max-leaf-certs 1 \
  --stores stores.csv

# Add certs listed in addCerts.csv to stores listed with no criteria. Then remove the certs listed in removeCerts.csv \
# from the stores.
kfutil stores rot reconcile \
  --add-certs addCerts.csv \
  --remove-certs removeCerts.csv \
  --stores stores.csv

# Remove all added certs from the demo
kfutil stores rot reconcile \
  --remove-certs removeCerts.csv \
  --stores stores.csv
kfutil stores rot reconcile \
  --remove-certs removeCerts2.csv \
  --stores stores.csv
kfutil stores rot reconcile \
  --remove-certs addCerts.csv \
  --stores stores.csv
kfutil stores rot reconcile \
  --remove-certs addCerts2.csv \
  --stores stores.csv

# List stores and convert to CSV into stores.csv format
echo '"StoreId","StoreType","StoreMachine","StorePath"' > meow.csv
kfutil stores list | jq -r '.[] | [.Id, .cert_store_type, .ClientMachine, .Storepath] |  @csv' >> meow.csv

