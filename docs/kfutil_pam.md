## kfutil pam

Keyfactor PAM Provider APIs.

### Synopsis

Privileged Access Management (PAM) functionality in Keyfactor Web APIs allows for configuration of third 
party PAM providers to secure certificate stores. The PAM component of the Keyfactor API includes methods necessary to 
programmatically create, delete, edit, and list PAM Providers.

### Options

```
  -h, --help   help for pam
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
* [kfutil pam create](kfutil_pam_create.md)	 - Create a new PAM Provider, currently only supported from file.
* [kfutil pam delete](kfutil_pam_delete.md)	 - Delete a defined PAM Provider by ID.
* [kfutil pam get](kfutil_pam_get.md)	 - Get a specific defined PAM Provider by ID.
* [kfutil pam list](kfutil_pam_list.md)	 - Returns a list of all the configured PAM providers.
* [kfutil pam types-create](kfutil_pam_types-create.md)	 - Creates a new PAM provider type.
* [kfutil pam types-list](kfutil_pam_types-list.md)	 - Returns a list of all available PAM provider types.
* [kfutil pam update](kfutil_pam_update.md)	 - Updates an existing PAM Provider, currently only supported from file.

###### Auto generated on 17-Jun-2025
