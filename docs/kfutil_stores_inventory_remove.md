## kfutil stores inventory remove

Removes a certificate from the certificate store inventory.

### Synopsis

Removes a certificate from the certificate store inventory.

```
kfutil stores inventory remove [flags]
```

### Options

```
      --all-stores           Remove the certificate(s) from all certificate stores.
      --cid strings          The Keyfactor command certificate ID(s) of the certificate to remove from the store(s).
      --client strings       Remove certificate(s) from all stores of specific client machine(s).
      --cn strings           Subject name(s) of the certificate(s) to remove from the store(s).
      --container strings    Remove certificate(s) from all stores of specific container type(s).
      --dry-run              Do not remove inventory, only show what would be removed.
      --force                Force removal of inventory without prompting for confirmation.
  -h, --help                 help for remove
      --sid strings          The Keyfactor Command ID of the certificate store(s) to remove inventory from.
      --store-type strings   Remove certificate(s) from all stores of specific store type(s).
      --thumbprint strings   The thumbprint of the certificate(s) to remove from the store(s).
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

* [kfutil stores inventory](kfutil_stores_inventory.md)	 - Commands related to certificate store inventory management

###### Auto generated on 17-Jun-2025
