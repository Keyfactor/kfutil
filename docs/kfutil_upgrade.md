## kfutil upgrade

Upgrade kfutil to the latest release

### Synopsis

Fetches a kfutil release from GitHub, verifies its SHA-256 checksum,
and atomically replaces the running binary.

By default the latest published release is installed. Pass --version with any
valid GitHub tag (e.g. v1.9.0, v1.10.0-beta.1) to install a specific release,
including pre-releases and older versions.

Examples:
  kfutil upgrade                      # install latest
  kfutil upgrade --version v1.8.0    # install a specific tag
  kfutil upgrade --dry-run           # preview without changing anything

```
kfutil upgrade [flags]
```

### Options

```
      --dry-run          Show what would be downloaded without replacing the binary
      --force            Upgrade even if already at the target version
  -h, --help             help for upgrade
      --version string   GitHub tag to install (default: latest release)
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
