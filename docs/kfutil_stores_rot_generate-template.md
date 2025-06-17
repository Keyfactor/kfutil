## kfutil stores rot generate-template

For generating Root Of Trust template(s)

### Synopsis

Root Of Trust: Will parse a CSV and attempt to deploy a cert or set of certs into a list of cert stores.

```
kfutil stores rot generate-template [flags]
```

### Options

```
      --cn strings               Subject name(s) to pre-populate the 'certs' template with. If not specified, the template will be empty. Does not work with SANs.
      --collection strings       Certificate collection name(s) to pre-populate the stores template with. If not specified, the template will be empty.
      --container-name strings   Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified container types. If not specified, the template will be empty.
  -f, --format csv               The type of template to generate. Only csv is supported at this time. (default "csv")
  -h, --help                     help for generate-template
  -o, --outpath string           Path to write the template file to. If not specified, the file will be written to the current directory.
      --store-type strings       Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified store types. If not specified, the template will be empty.
      --type string              The type of template to generate. Only "certs|stores|actions" are supported at this time. (default "certs")
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
