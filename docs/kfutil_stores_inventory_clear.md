## kfutil stores inventory clear

Clears the certificate store store inventory of ALL certificates.

### Synopsis

Clears the certificate store store inventory of ALL certificates.

```
kfutil stores inventory clear [flags]
```

### Options

```
      --all                 Remove all inventory from all certificate stores.
      --client strings      Remove all inventory from store(s) of specific client machine(s).
      --container strings   Remove all inventory from store(s) of specific container type(s).
      --dry-run             Do not remove inventory, only show what would be removed.
      --force               Force removal of inventory without prompting for confirmation.
  -h, --help                help for clear
      --sid strings         The Keyfactor Command ID of the certificate store(s) remove all inventory from.
      --store-ype strings   Remove all inventory from store(s) of specific store type(s).
```

### SEE ALSO

* [kfutil stores inventory](kfutil_stores_inventory.md)	 - Commands related to certificate store inventory management

###### Auto generated by spf13/cobra on 1-Dec-2022
