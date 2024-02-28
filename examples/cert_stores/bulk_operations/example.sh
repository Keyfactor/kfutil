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

#  list_stores
#  generate_store_template
  export_stores
  import_stores
}

function export_stores() {
  kfutil stores export
}

function import_stores(){
  # Directory containing the CSV files
  DIRECTORY=$(pwd)

  # Find all CSV files in the specified directory
  CSV_FILES=($(find "$DIRECTORY" -name "*.csv"))

  # Check if no CSV files were found
  if [ ${#CSV_FILES[@]} -eq 0 ]; then
      echo "No CSV files found in the directory."
      exit 1
  fi

  # Display the CSV files to the user
  echo "Select a CSV file by number:"
  for i in "${!CSV_FILES[@]}"; do
      echo "$((i+1))) ${CSV_FILES[$i]}"
  done

  # Prompt the user for a choice
  read -p "Enter number (1-${#CSV_FILES[@]}): " USER_CHOICE

  # Validate the user input
  if ! [[ "$USER_CHOICE" =~ ^[0-9]+$ ]] || [ "$USER_CHOICE" -lt 1 ] || [ "$USER_CHOICE" -gt ${#CSV_FILES[@]} ]; then
      echo "Invalid selection. Please run the script again and select a valid number."
      exit 1
  fi

  # Calculate the index of the selected file
  SELECTED_INDEX=$((USER_CHOICE-1))

  # Get the selected CSV file
  SELECTED_CSV_FILE=${CSV_FILES[$SELECTED_INDEX]}

  # Execute the command with the selected CSV file
  kfutil stores import csv \
    --file="$SELECTED_CSV_FILE"

  echo "CSV import command executed for file: $SELECTED_CSV_FILE"
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
  command_str="kfutil stores import generate-template"

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

function cleanup() {
  echo "Cleaning up..."
  unset outpath
  unset KFUTIL_STORE_TEMPLATE_PATH
  rm -f *.csv
  echo "Cleanup complete."
}

prompt_user
cleanup

