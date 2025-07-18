## kfutil helm uo

Configure the Keyfactor Universal Orchestrator Helm Chart

### Synopsis

Configure the Keyfactor Universal Orchestrator Helm Chart by prompting the user for configuration values and outputting a YAML file that can be used with the Helm CLI to install the chart.

Also supported is the ability specify extensions and skip the interactive prompts.


```
kfutil helm uo [-t <token>] [-o <path>] [-f <file, url, or '-'>] [-e <extension name>@<version>]... [flags]
```

### Options

```
  -e, --extension strings   List of extensions to install. Should be in the format <extension name>@<version>. If no version is specified, the latest version will be downloaded.
  -h, --help                help for uo
  -o, --out string          Path to output the modified values.yaml file. This file can then be used with helm install -f <file> to override the default values.
  -t, --token string        Token used for related authentication - required for private repositories
  -f, --values strings      Filename, directory, or URL to a default values.yaml file to use for the chart
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

* [kfutil helm](kfutil_helm.md)	 - Helm utilities for configuring Keyfactor Helm charts

###### Auto generated on 17-Jun-2025
