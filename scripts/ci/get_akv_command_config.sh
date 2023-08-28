#!/usr/bin/env bash
set -e -o pipefail

# Default values
keyvault_name=""
secret_name=""

# Check if environmental variables are set
if [ -n "$KFUTIL_KEYVAULT_NAME" ]; then
    keyvault_name="$KFUTIL_KEYVAULT_NAME"
fi

if [ -n "$KFUTIL_KEYVAULT_SECRET_NAME" ]; then
    secret_name="$KFUTIL_KEYVAULT_SECRET_NAME"
fi

# Parse command line options
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --keyvault)
            keyvault_name="$2"
            shift
            shift
            ;;
        --secret)
            secret_name="$2"
            shift
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if az command is available
if ! command -v az &> /dev/null; then
    echo "Azure CLI (az) is not installed. Please install it and try again."
    exit 1
fi

# Check if keyvault name and secret name are provided
if [ -z "$keyvault_name" ] || [ -z "$secret_name" ]; then
    echo "Usage: $0 [--keyvault <keyvault_name>] [--secret <secret_name>]"
    exit 1
fi

# Retrieve the secret using Azure CLI
az keyvault secret show --vault-name "$keyvault_name" --name "$secret_name" --output json | jq -r .value | jq -r > "${HOME}/.keyfactor/command_config.json"
