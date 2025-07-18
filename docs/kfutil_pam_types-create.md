## kfutil pam types-create

Creates a new PAM provider type.

### Synopsis

Creates a new PAM Provider type, currently only supported from JSON file and from GitHub. To install from 
Github. To install from GitHub, use the --repo flag to specify the GitHub repository and optionally the branch to use. 
NOTE: the file from Github must be named integration-manifest.json and must use the same schema as 
https://github.com/Keyfactor/hashicorp-vault-pam/blob/main/integration-manifest.json. To install from a local file, use
--from-file to specify the path to the JSON file.

```
kfutil pam types-create [flags]
```

### Options

```
  -b, --branch string      Branch name for the repository. Defaults to 'main'.
  -f, --from-file string   Path to a JSON file containing the PAM Type Object Data.
  -h, --help               help for types-create
  -n, --name string        Name of the PAM Provider Type.
  -r, --repo string        Keyfactor repository name of the PAM Provider Type.
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

* [kfutil pam](kfutil_pam.md)	 - Keyfactor PAM Provider APIs.

###### Auto generated on 17-Jun-2025
