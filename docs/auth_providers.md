# Auth Providers
What is an `auth provider` in the conext of `kfutil`? It's a way to source credentials needed to connect to a Keyfactor
product or service from a secure location rather than a file on disk or environment variables. 

## Available Auth Providers
- [Azure Key Vault](#azure-key-vault)

## Azure Key Vault
The Azure Key Vault auth provider allows you to source credentials from an Azure Key Vault instance using Azure Managed
Identity. 

### Configuration
Below is an example configuration for the Azure Key Vault auth provider. This can be placed in the `$HOME/.keyfactor/command_config.json`
file and will be used by `kfutil` to source credentials for the Keyfactor product or service you are connecting to.
```json
{
  "servers": {
    "default": {
      "auth_provider": {
        "type": "azid",
        "profile": "default",
        "parameters": {
          "secret_name": "command-config-1021",
          "vault_name": "kfutil"
        }
      }
    }
  }
}
```
#### Usage

##### Default
With the above configuration in placed in the default path `$HOME/.keyfactor/command_config.json` the utility will 
implicitly attempt to source credentials from the Azure Key Vault instance.
```bash
kfutil stores list
```

##### Explicit
You can also explicitly specify the auth provider to use by passing the `--auth-provider` flags to the utility as shown
below. The file format will still be the same as above.
```bash
kfutil \
  --auth-provider-type azid \
  --auth-provider-profile default \
  --config /path/to/config/file.json \
  stores list
```
