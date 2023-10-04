#!/usr/bin/env bash
set -e -o pipefail

# Define the default values using environment variables
default_vault_name="${VAULT_NAME:-kfutil}"
default_secret_name="${SECRET_NAME:-integration-labs}"
echo "Default vault name: $default_vault_name"
echo "Default secret name: $default_secret_name"

export METADATA_URL="http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://vault.azure.net"


read_keyvault_secret() {
  local vault_name="$1"
  local secret_name="$2"

  echo "Vault Name: $vault_name"
  echo "Secret Name: $secret_name"

  # Make a request to the metadata endpoint
  echo "Querying metadata endpoint for access token..."
  echo "Metadata URL: $METADATA_URL"
  token_json=$(curl -H "Metadata: true" $METADATA_URL)

  echo "Exporting access token to access_token variable..."
  # Parse the access token from the response JSON
  access_token=$(echo $token_json | jq -r .access_token)

  # Now you can use the $access_token to authenticate and access Azure Key Vault
  echo "Access Token: $access_token"

  secret_url="https://${vault_name}.vault.azure.net/secrets/${secret_name}?api-version=7.0"
  echo "Secret URL: $secret_url"

  # Create a new secret in Azure Key Vault
  #secret_value="meow"
  #curl -X PUT -H "Authorization: Bearer ${access_token}" -H "Content-Type: application/json" -d "{\"value\": \"${secret_value}\"}" "$secret_url"

  # Get the secret value from Azure Key Vault
  echo "Querying Azure Key Vault for secret value..."
  secret_value=$(curl -H "Authorization: Bearer ${access_token}" "$secret_url" | jq -r .value)

  #echo "Secret Value: $secret_value"
  mkdir -p ~/.keyfactor
  echo "${secret_value}" | jq -r . > "${secret_name}.json"
  rm -f "${HOME}/.keyfactor/command_config.json" || true
  echo "${secret_value}" | jq -r . > "${HOME}/.keyfactor/command_config.json"
#  echo $secret_value > .env
}

# Main script logic
if [[ $# -eq 0 ]]; then
    # No arguments provided, use default values from environment variables
    read_keyvault_secret "$default_vault_name" "$default_secret_name"
elif [[ $# -eq 2 ]]; then
    # Two arguments provided: vault_name and secret_name
    read_keyvault_secret "$1" "$2"
else
    echo "Usage: $0 [vault_name secret_name]"
    exit 1
fi
