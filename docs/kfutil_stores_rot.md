## kfutil stores rot

Root of trust utility

### Synopsis

Root of trust allows you to manage your trusted roots using Keyfactor certificate stores.
For example if you wish to add a list of "root" certs to a list of certificate stores you would simply generate and fill
out the template CSV file. These template files can be generated with the following commands:
kfutil stores rot generate-template --type certs
kfutil stores rot generate-template --type stores
Once those files are filled out you can use the following command to add the certs to the stores:
kfutil stores rot audit --certs-file <certs-file> --stores-file <stores-file>
Will generate a CSV report file 'rot_audit.csv' of what actions will be taken. If those actions are correct you can run
the following command to actually perform the actions:
kfutil stores rot reconcile --certs-file <certs-file> --stores-file <stores-file>
OR if you want to use the audit report file generated you can run this command:
kfutil stores rot reconcile --import-csv <audit-file>


### Options

```
  -h, --help   help for rot
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

* [kfutil stores](kfutil_stores.md)	 - Keyfactor certificate stores APIs and utilities.
* [kfutil stores rot audit](kfutil_stores_rot_audit.md)	 - Audit generates a CSV report of what actions will be taken based on input CSV files.
* [kfutil stores rot generate-template](kfutil_stores_rot_generate-template.md)	 - For generating Root Of Trust template(s)
* [kfutil stores rot reconcile](kfutil_stores_rot_reconcile.md)	 - Reconcile either takes in or will generate an audit report and then add/remove certs as needed.

###### Auto generated on 17-Jun-2025
