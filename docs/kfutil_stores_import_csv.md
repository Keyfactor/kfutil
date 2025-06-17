## kfutil stores import csv

Create certificate stores from CSV file.

### Synopsis

Will parse a CSV file and attempt to create a certificate store for each row with the provided parameters.
Any errors encountered will be logged to the <file_name>_results.csv file, under the 'Errors' column.

Required Flags:
- '--store-type-name' OR '--store-type-id'
- '--file' is the path to the file to be imported.

#### Credentials

##### In the CSV file:

###### Credential Fields

| Header                    | Description                                                                           |
|---------------------------|---------------------------------------------------------------------------------------|
| Properties.ServerUsername | This is equivalent to the 'ServerUsername' field in the Command Certificate Store UI. |
| Properties.ServerPassword | This is equivalent to the 'ServerPassword' field in the Command Certificate Store UI. |
| Password                  | This is equivalent to the 'StorePassword' field in the Command Certificate Store UI.  |

###### Inventory Schedule Fields

| Header                             | Description                                                                                    |
|------------------------------------|------------------------------------------------------------------------------------------------|
| InventorySchedule.Immediate        | Boolean value, set to "TRUE" to schedule immediate inventory                                   |	
| InventorySchedule.Interval.Minutes | The timeframe in which to periodically inventory int number/integer value. Ex.120 for 2 hours. |
| InventorySchedule.Daily.Time       | The time of day to inventory daily, RFC3339 format. Ex. "2023-10-01T12:00:00Z" for noon UTC.   |	
| InventorySchedule.Weekly.Days      | TBD                                                                                            |	
| InventorySchedule.Weekly.Time      | The time of day to inventory daily, RFC3339 format. Ex. "2023-10-01T12:00:00Z" for noon UTC.   |

##### Outside CSV file:
If you do not wish to include credentials in your CSV file they can be provided one of three ways:
- via the --server-username --server-password and --store-password flags
- via environment variables: KFUTIL_CSV_SERVER_USERNAME, KFUTIL_CSV_SERVER_PASSWORD, KFUTIL_CSV_STORE_PASSWORD
- via interactive prompts


```
kfutil stores import csv --file <file name to import> --store-type-id <store type id> --store-type-name <store type name> --results-path <filepath for results> --dry-run <check fields only> [flags]
```

### Options

```
  -d, --dry-run                                     Do not import, just check for necessary fields.
  -f, --file string                                 CSV file containing cert stores to create.
  -h, --help                                        help for csv
  -o, --results-path string                         CSV file containing cert stores to create. defaults to <imported file name>_results.csv
  -p, --server-password Properties.ServerPassword   The password Keyfactor Command will use to use connect to the certificate store host. This field can be specified in the CSV file in the column Properties.ServerPassword. This value can also be sourced from the environmental variable `KFUTIL_CSV_SERVER_PASSWORD`. *NOTE* a value provided in the CSV file will override any other input value
  -u, --server-username Properties.ServerUsername   The username Keyfactor Command will use to use connect to the certificate store host. This field can be specified in the CSV file in the column Properties.ServerUsername. This value can also be sourced from the environmental variable `KFUTIL_CSV_SERVER_USERNAME`. *NOTE* a value provided in the CSV file will override any other input value
  -s, --store-password Password                     The credential information Keyfactor Command will use to access the certificates in a specific certificate store (the store password). This is different from credential information Keyfactor Command uses to access a certificate store host. This field can be specified in the CSV file in the column Password. This value can also be sourced from the environmental variable `KFUTIL_CSV_STORE_PASSWORD`. *NOTE* a value provided in the CSV file will override any other input value
  -i, --store-type-id int                           The ID of the cert store type for the stores. (default -1)
  -n, --store-type-name string                      The name of the cert store type.  Use if store-type-id is unknown.
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

* [kfutil stores import](kfutil_stores_import.md)     - Import a file with certificate store definitions and create them
  in Keyfactor Command.

###### Auto generated on 17-Jun-2025
