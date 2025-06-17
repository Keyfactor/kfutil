## kfutil stores inventory clear

Clears the certificate store store inventory of ALL certificates.

### Synopsis

Clears the certificate store store inventory of ALL certificates.

```
kfutil stores inventory clear [flags]
```

### Options

```
      --all                  Remove all inventory from all certificate stores.
      --client strings       Remove all inventory from store(s) of specific client machine(s).
      --container strings    Remove all inventory from store(s) of specific container type(s).
      --dry-run              Do not remove inventory, only show what would be removed.
      --force                Force removal of inventory without prompting for confirmation.
  -h, --help                 help for clear
      --sid strings          The Keyfactor Command ID of the certificate store(s) remove all inventory from.
      --store-type strings   Remove all inventory from store(s) of specific store type(s).
```

### Options inherited from parent commands

```
      --config string    Full path to config file in JSON format. (default is $HOME/.keyfactor/command_config.json)
      --debug            Enable debug logging. (USE AT YOUR OWN RISK, this may log sensitive information to the console.)
      --exp              Enable experimental features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)
      --no-prompt        Do not prompt for any user input and assume defaults or environmental variables are set.
      --profile string   Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.
```

### SEE ALSO

* [kfutil stores inventory](kfutil_stores_inventory.md)	 - Commands related to certificate store inventory management

###### Auto generated on 14-Jun-2023
