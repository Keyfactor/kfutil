## kfutil stores

Keyfactor certificate stores APIs and utilities.

### Synopsis

A collections of APIs and utilities for interacting with Keyfactor certificate stores.

### Options

```
  -h, --help   help for stores
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
* [kfutil stores delete](kfutil_stores_delete.md)	 - Delete a certificate store by ID.
* [kfutil stores export](kfutil_stores_export.md)	 - Export existing defined certificate stores by type or store Id.
* [kfutil stores get](kfutil_stores_get.md)	 - Get a certificate store by ID.
* [kfutil stores import](kfutil_stores_import.md)     - Import a file with certificate store definitions and create them
  in Keyfactor Command.
* [kfutil stores inventory](kfutil_stores_inventory.md)	 - Commands related to certificate store inventory management
* [kfutil stores list](kfutil_stores_list.md)	 - List certificate stores.
* [kfutil stores rot](kfutil_stores_rot.md)     - Root of trust utility

###### Auto generated on 17-Jun-2025
