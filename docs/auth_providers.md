# Auth Providers

What is an `auth provider` in the context of `kfutil`? It's a way to source credentials needed to connect to a Keyfactor
Command API from a secure location rather than a file on disk or environment variables.

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
          "secret_name": "kfutil-credentials",
          "vault_name": "keyfactor-command-secrets"
        }
      }
    }
  }
}
```

### Azure Key Vault Secret Format
The format of the Azure Key Vault secret should be the same as if you were to run `kfutil login` and go through the
interactive auth flow. Here's an example of what that would look like:

#### Basic Auth Example
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

#### oAuth Client Credentials Example

```json
{
  "servers": {
    "default": {
      "host": "my.kfcommand.domain",
      "client_id": "my_oauth_client_id",
      "client_secret": "my_oauth_client_secret",
      "token_url": "https://my_oauth_token_url",
      "api_path": "Keyfactor/API"
    }
  }
}
```

#### oAuth Client Credentials with a Third-party IdP (Microsoft Entra ID / Azure AD)

`kfutil` can authenticate against any OAuth2 IdP that Keyfactor Command trusts, not only the Keyfactor-hosted
Keycloak. The example below configures **Microsoft Entra ID** (formerly Azure AD) using the client-credentials
grant. The Entra-side setup — registering the Command API application, exposing its API / Application ID URI,
and granting the calling app permission — lives in the Keyfactor Command / Entra SSO documentation and is a
prerequisite.

```json
{
  "servers": {
    "default": {
      "host": "keyfactor.example.com",
      "auth_type": "oauth",
      "client_id": "<entra-app-client-id>",
      "client_secret": "<entra-client-secret>",
      "token_url": "https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token",
      "scopes": ["api://<command-app-id-uri>/.default"],
      "api_path": "KeyfactorAPI"
    }
  }
}
```

Key points when targeting Entra ID:

- **`token_url`** is the Entra v2.0 token endpoint for your tenant:
  `https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token`.
- **`client_id`** / **`client_secret`** are the Entra app registration's *Application (client) ID* and a
  generated *client secret*.
- **`scopes`** is **mandatory** and must be `api://<command-app-id-uri>/.default`. `kfutil` sends no scope by
  default, so an Entra request with no scope configured is rejected.
- **`audience`** must be **omitted**. Entra ID does not accept an `audience` form parameter.

##### Scope vs. audience

This is the most common point of confusion when moving from Keycloak to Entra ID:

- **Keycloak** (the Keyfactor-hosted IdP) identifies the target resource with the `audience` field, and scopes
  are typically optional.
- **Entra ID** identifies the target resource through the requested **scope** (`api://<command-app-id-uri>/.default`)
  and does **not** support an `audience` parameter.

Therefore, for Entra ID set `scopes` and leave `audience` unset. Setting `audience` against Entra will cause the
token request to fail. Note also that the `KEYFACTOR_AUTH_SCOPES` environment variable is split on `,` (comma),
so when using env vars supply exactly one Entra scope with no commas.

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
