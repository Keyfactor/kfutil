## kfutil store-types

Keyfactor certificate store types APIs and utilities.

### Synopsis

A collections of APIs and utilities for interacting with Keyfactor certificate store types.

### Options

```
  -h, --help   help for store-types
```

### Options inherited from parent commands

```
      --api-path string                API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI) (default "KeyfactorAPI")
      --auth-provider-profile string   The profile to use defined in the securely stored config. If not specified the config named 'default' will be used if it exists. (default "default")
      --auth-provider-type string      Provider type choices: (azid)
      --client-id string               OAuth2 client-id to use for authenticating to Keyfactor Command.
      --client-secret string           OAuth2 client-secret to use for authenticating to Keyfactor Command.
      --config string                  Full path to config file in JSON format. (default is $HOME/.keyfactor/command_config.json)
      --debug                          Enable debugFlag logging.
      --domain string                  Domain to use for authenticating to Keyfactor Command.
      --exp                            Enable expEnabled features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)
      --format text                    How to format the CLI output. Currently only text is supported. (default "text")
      --hostname string                Hostname to use for authenticating to Keyfactor Command.
      --no-prompt                      Do not prompt for any user input and assume defaults or environmental variables are set.
      --offline                        Will not attempt to connect to GitHub for latest release information and resources.
      --password string                Password to use for authenticating to Keyfactor Command. WARNING: Remember to delete your console history if providing kfcPassword here in plain text.
      --profile string                 Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.
      --skip-tls-verify                Disable TLS verification for API requests to Keyfactor Command.
      --token-url string               OAuth2 token endpoint full URL to use for authenticating to Keyfactor Command.
      --username string                Username to use for authenticating to Keyfactor Command.
```

### SEE ALSO

* [kfutil](kfutil.md)	 - Keyfactor CLI utilities
* [kfutil store-types create](kfutil_store-types_create.md)	 - Create a new certificate store type in Keyfactor.
* [kfutil store-types delete](kfutil_store-types_delete.md)	 - Delete a specific store type by name or ID.
* [kfutil store-types get](kfutil_store-types_get.md)	 - Get a specific store type by either name or ID.
* [kfutil store-types list](kfutil_store-types_list.md)	 - List certificate store types.
* [kfutil store-types templates-fetch](kfutil_store-types_templates-fetch.md)	 - Fetches store type templates from Keyfactor's Github.

###### Auto generated on 17-Jun-2025
