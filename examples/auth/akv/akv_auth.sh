#!/usr/bin/env bash
set -e -o pipefail

export VAULT_NAME="kfutil"
export SECRET_NAME="integration-labs"

export METADATA_URL="http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://vault.azure.net"
echo "Metadata URL: $METADATA_URL"

# Make a request to the metadata endpoint
echo "Querying metadata endpoint for access token..."
TOKEN_JSON=$(curl -H "Metadata: true" $METADATA_URL)

echo "Exporting access token to ACCESS_TOKEN variable..."
# Parse the access token from the response JSON
export ACCESS_TOKEN=$(echo $TOKEN_JSON | jq -r .access_token)

# Now you can use the $ACCESS_TOKEN to authenticate and access Azure Key Vault
echo "Access Token: $ACCESS_TOKEN"

export SECRET_URL="https://${VAULT_NAME}.vault.azure.net/secrets/${SECRET_NAME}?api-version=7.0"
echo "Secret URL: $SECRET_URL"

# Create a new secret in Azure Key Vault
#SECRET_VALUE="meow"
#curl -X PUT -H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json" -d "{\"value\": \"${SECRET_VALUE}\"}" "$SECRET_URL"

# Get the secret value from Azure Key Vault
echo "Querying Azure Key Vault for secret value..."
SECRET_VALUE=$(curl -H "Authorization: Bearer ${ACCESS_TOKEN}" "$SECRET_URL" | jq -r .value)

#echo "Secret Value: $SECRET_VALUE"
mkdir -p ~/.keyfactor
echo $SECRET_VALUE | jq -r . > ~/.keyfactor/command_config.json
