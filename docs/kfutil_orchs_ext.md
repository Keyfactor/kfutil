## kfutil orchs ext

Download and configure extensions for Keyfactor Command Universal Orchestrator

### Synopsis


Keyfactor Command Universal Orchestrator utility for downloading and configuring extensions.

This command will download extensions for Keyfactor Command Universal Orchestrator. Extensions can be downloaded from a configuration file or by specifying the extension name and version.


```
kfutil orchs ext [-t <token>] [--org <Github org>] [-o <out path>] [-c <config file> | -e <extension name>@<version>] [-y] [-u] [-P] [flags]
```

### Examples

```
ext -t <token> -e <extension>@<version>,<extension>@<version> -o ./app/extensions --confirm, --update, --prune
```

### Options

```
  -c, --config strings      Filename, directory, or URL to an extension configuration file to use for the extension
  -y, --confirm             Automatically confirm the download of extensions
  -e, --extension strings   List of extensions to download. Should be in the format <extension name>@<version>. If no version is specified, the latest official version will be downloaded.
  -h, --help                help for ext
      --org string          Github organization to download extensions from. Default is keyfactor.
  -o, --out string          Path to the extensions directory to download extensions into. Default is ./extensions
  -P, --prune               Remove extensions from the extensions directory that are not in the extension configuration file or specified on the command line
  -t, --token string        Token used for related authentication - required for private repositories
  -u, --update              Update existing extensions if they are out of date.
```

### Options inherited from parent commands

```
      --api-path string                API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI) (default "KeyfactorAPI")
      --auth-provider-profile string   The profile to use defined in the securely stored config. If not specified the config named 'default' will be used if it exists. (default "default")
      --auth-provider-type string      Provider type choices: (azid)
      --client-id string               OAuth2 client-id to use for authenticating to Keyfactor Command.
      --client-secret string           OAuth2 client-secret to use for authenticating to Keyfactor Command.
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

* [kfutil orchs](kfutil_orchs.md)	 - Keyfactor agents/orchestrators APIs and utilities.

###### Auto generated on 17-Jun-2025
