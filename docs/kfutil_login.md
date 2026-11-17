## kfutil login

User interactive login to Keyfactor. Stores the credentials in the config file '$HOME/.keyfactor/command_config.json'.

### Synopsis

Will prompt the user for a Username and Password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for Username and Password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables.

For more information on the environment variables review the docs: https://github.com/Keyfactor/kfutil/tree/main?tab=readme-ov-file#environmental-variables

WARNING: This will write the environmental credentials to disk and will be stored in the config file in plain text at:
'$HOME/.keyfactor/command_config.json.'


```
kfutil login [flags]
```

### Options

```
  -h, --help            help for login
      --skip-validate   Save the login configuration without validating credentials against Keyfactor Command.
```

### Options inherited from parent commands

```
      --api-path string                API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI) (default "KeyfactorAPI")
      --audience string                OAuth2 audience to request when authenticating to Keyfactor Command. Not used by all IdPs (e.g. leave unset for Entra ID).
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
      --scopes string                  Comma-separated list of OAuth2 scopes to request when authenticating via a third-party IdP (e.g. Entra ID: api://<app-id-uri>/.default).
      --skip-tls-verify                Disable TLS verification for API requests to Keyfactor Command.
      --token-url string               OAuth2 token endpoint full URL to use for authenticating to Keyfactor Command.
      --username string                Username to use for authenticating to Keyfactor Command.
```

### SEE ALSO

* [kfutil](kfutil.md)	 - Keyfactor CLI utilities
