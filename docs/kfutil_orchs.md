## kfutil orchs

Keyfactor agents/orchestrators APIs and utilities.

### Synopsis

A collections of APIs and utilities for interacting with Keyfactor orchestrators.

### Options

```
  -h, --help   help for orchs
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
* [kfutil orchs approve](kfutil_orchs_approve.md)	 - Approve orchestrator by machine/client name.
* [kfutil orchs disapprove](kfutil_orchs_disapprove.md)	 - Disapprove orchestrator by machine/client name.
* [kfutil orchs ext](kfutil_orchs_ext.md)	 - Download and configure extensions for Keyfactor Command Universal Orchestrator
* [kfutil orchs get](kfutil_orchs_get.md)	 - Get orchestrator by machine/client name.
* [kfutil orchs list](kfutil_orchs_list.md)	 - List orchestrators.
* [kfutil orchs logs](kfutil_orchs_logs.md)	 - Get orchestrator logs by machine/client name.
* [kfutil orchs reset](kfutil_orchs_reset.md)	 - Reset orchestrator by machine/client name.

###### Auto generated on 17-Jun-2025
