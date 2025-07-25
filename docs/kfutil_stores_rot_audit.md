## kfutil stores rot audit

Audit generates a CSV report of what actions will be taken based on input CSV files.

### Synopsis

Root of Trust Audit: Will read and parse inputs to generate a report of certs that need to be added or removed from the "root of trust" stores.

```
kfutil stores rot audit [flags]
```

### Options

```
  -a, --add-certs string      CSV file containing cert(s) to enroll into the defined cert stores
  -d, --dry-run               Dry run mode
  -h, --help                  help for audit
  -k, --max-keys -1           The max number of private keys that should be in a store to be considered a 'root' store. If set to -1 then all stores will be considered. (default -1)
  -l, --max-leaf-certs -1     The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to -1 then all stores will be considered. (default -1)
  -m, --min-certs -1          The minimum number of certs that should be in a store to be considered a 'root' store. If set to -1 then all stores will be considered. (default -1)
  -o, --outpath string        Path to write the audit report file to. If not specified, the file will be written to the current directory.
  -r, --remove-certs string   CSV file containing cert(s) to remove from the defined cert stores
  -s, --stores string         CSV file containing cert stores to enroll into
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

* [kfutil stores rot](kfutil_stores_rot.md)	 - Root of trust utility

###### Auto generated on 17-Jun-2025
