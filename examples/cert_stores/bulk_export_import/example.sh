#!/usr/bin/env bash
set -e -o pipefail

function prompt_user() {
  echo "This script will install kfutil and generate a store template."
  read -p "Do you want to continue? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Exiting..."
    exit 1
  fi

  install_kfutil

  # Prompt to create a new certificate store type
  read -p "Do you want to create a new certificate store type? (y/n) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    kfutil store-types create
    echo "Please (re)register and approve an orchestrator that supports this newly created store type and rerun this script."
    exit 1
  fi

  list_stores
  generate_store_template

}

function list_stores() {
  echo "Checking that at least one certificate store exists to export..."
  echo kfutil stores list
  # Check if output was [] or if there was exit code non-zero
  if [[ $(kfutil stores list) -ne 0 || $(kfutil stores list) == "[]" ]]; then
    echo "No certificate stores found. At least one is required to export."
    exit 1
  fi
  kfutil stores list
  return
}

function install_kfutil() {
  # Check if `kfutil` is installed
  if command -v kfutil &> /dev/null; then
    echo "kfutil is installed"
    kfutil version
    echo
    return
  else
    bash <(curl -s https://raw.githubusercontent.com/Keyfactor/kfutil/main/install.sh)
  fi
}

function generate_store_template() {
  local store_type_id="$1"
  local store_type_name="$2"
  local outpath="$3"
  command_str="kfutil stores import generate-template"

  # If neither store_type_id nor store_type_name is provided, exit
  if [[ -z "$store_type_id" && -z "$store_type_name" ]]; then
    # Prompt for an ID or name
    read -p "Please provide a store type ID or name: " store_type_id
    # Check if input is a number
    if [[ $store_type_id =~ ^[0-9]+$ ]]; then
      # If input is a number, use it as store_type_id
      store_type_name=""
      # append store_type_id to command_str
      command_str="$command_str --store-type-id $store_type_id"
    else
      store_type_name="$store_type_id"
      # append store_type_name to command_str
      command_str="$command_str --store-type-name $store_type_name"
    fi
  fi

  # If outpath is not provided, check for environment variable
  if [[ -z "$outpath" ]]; then
    if [[ -z "$KFUTIL_STORE_TEMPLATE_PATH" ]]; then
      echo "Template file will be saved to the current working directory."
    else
      # append environment variable to command_str
      command_str="$command_str --outpath $KFUTIL_STORE_TEMPLATE_PATH"
    fi
  else
    # append outpath to command_str
    command_str="$command_str --outpath $outpath"
  fi
  echo "Generating store template file"
  echo
  echo "$command_str"
  eval "$command_str"
}

prompt_user

