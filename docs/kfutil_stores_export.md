## kfutil stores export

Export existing defined certificate stores by type or store Id.

### Synopsis

Export the parameter values of defined certificate stores either by type or a specific store by Id. These parameters are stored in CSV for importing later.

```
kfutil stores export [flags]
```

### Options

```
  -a, --all                      Export all stores grouped by store-type.
  -h, --help                     help for export
  -o, --outpath string           Path and name of the template file to generate.. If not specified, the file will be written to the current directory.
  -i, --store-type-id int        The ID of the cert store type for the template. (default -1)
  -n, --store-type-name string   The name of the cert store type for the template.  Use if store-type-id is unknown.
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

* [kfutil stores](kfutil_stores.md)	 - Keyfactor certificate stores APIs and utilities.

###### Auto generated on 24-Apr-2025
