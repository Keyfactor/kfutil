## kfutil import

Keyfactor instance import utilities.

### Synopsis

A collection of APIs and utilities for importing Keyfactor instance data.

```
kfutil import [flags]
```

### Options

```
  -a, --all                    import all importable data to JSON file
  -c, --collections            import collections to JSON file
  -d, --denied-alerts          import denied cert alerts to JSON file
  -f, --file string            path to JSON file containing exported data
  -h, --help                   help for import
  -i, --issued-alerts          import issued cert alerts to JSON file
  -m, --metadata               import metadata to JSON file
  -n, --networks               import SSL networks to JSON file
  -p, --pending-alerts         import pending cert alerts to JSON file
  -r, --reports                import reports to JSON file
  -s, --security-roles         import security roles to JSON file
  -w, --workflow-definitions   import workflow definitions to JSON file
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

###### Auto generated on 17-Jun-2025
