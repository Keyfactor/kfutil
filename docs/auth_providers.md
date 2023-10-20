# Auth Providers
What is an `auth provider` in the conext of `kfutil`? It's a way to source credentials needed to connect to a Keyfactor
product or service from a secure location rather than a file on disk or environment variables. 

* [Available Auth Providers](#available-auth-providers)
* [Azure Key Vault](#azure-key-vault)
    + [Configuration](#configuration)
    + [Azure Key Vault Secret Format](#azure-key-vault-secret-format)
        - [Usage](#usage)
            * [Default](#default)
            * [Explicit](#explicit)

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

### Azure Key Vault Secret Format
The format of the Azure Key Vault secret should be the same as if you were to run `kfutil login` and go through the 
interactive auth flow. Here's an example of what that would look like:
```json
{
  "servers": {
    "default": {
      "host": "my.kfcommand.domain",
      "username": "my_kfcommand_username",
      "password": "my_kfcommand_password",
      "domain": "my_kfcommand_domain",
      "api_path": "KeyfactorAPI"
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
The above explicitly tells the utility to only attempt to use the Azure Key Vault auth provider. This mode will not fail
to user interactive or environmental variable auth if provided. The example also shows how to specify a custom path to
the auth provider configuration file and what profile to look for in the configuration file stored in Azure.