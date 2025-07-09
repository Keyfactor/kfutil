## kfutil stores inventory add

Adds one or more certificates to one or more certificate store inventories.

### Synopsis

Adds one or more certificates to one or more certificate store inventories. The certificate(s) to add can be
specified by thumbprint, Keyfactor command certificate ID, or subject name. The store(s) to add the certificate(s) to can be
specified by Keyfactor command store ID, client machine name, store type, or container type. At least one or more stores
and one or more certificates must be specified. If multiple stores and/or certificates are specified, the command will
attempt to add all the certificate(s) meeting the specified criteria to all stores meeting the specified criteria.

```
kfutil stores inventory add [flags]
```

### Options

```
      --all-stores           Add the certificate(s) to all certificate stores.
      --cid strings          The Keyfactor command certificate ID(s) of the certificate to add to the store(s).
      --client strings       Add a certificate to all stores of specific client machine(s).
      --cn strings           Subject name(s) of the certificate(s) to add to the store(s).
      --container strings    Add a certificate to all stores of specific container type(s).
      --dry-run              Do not add inventory, only show what would be added.
      --force                Force addition of inventory without prompting for confirmation.
  -h, --help                 help for add
      --sid strings          The Keyfactor Command ID of the certificate store(s) to add inventory to.
      --store-type strings   Add a certificate to all stores of specific store type(s).
      --thumbprint strings   The thumbprint of the certificate(s) to add to the store(s).
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
