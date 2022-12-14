## kfutil stores inventory add

Adds one or more certificates to one or more certificate store inventories.

### Synopsis

Adds one or more certificates to one or more certificate store inventories. The certificate(s) to add can be
specified by thumbprint, Keyfactor command certificate ID, or subject name. The store(s) to add the certificate(s) to can be
specified by Keyfactor command store ID, client machine name, store type, or container type. At least one or more stores
and one or more certificates must be specified. If multiple stores and/or certificates are specified, the command will
attempt to add all the certificate(s) meeting the specified criteria to all stores meeting the specified criteria.

```
kfutil stores inventory add [flags]
```

### Options

```
      --all-stores           Add the certificate(s) to all certificate stores.
      --cid strings          The Keyfactor command certificate ID(s) of the certificate to add to the store(s).
      --client strings       Add a certificate to all stores of specific client machine(s).
      --cn strings           Subject name(s) of the certificate(s) to add to the store(s).
      --container strings    Add a certificate to all stores of specific container type(s).
      --dry-run              Do not add inventory, only show what would be added.
      --force                Force addition of inventory without prompting for confirmation.
  -h, --help                 help for add
      --sid strings          The Keyfactor Command ID of the certificate store(s) to add inventory to.
      --store-type strings   Add a certificate to all stores of specific store type(s).
      --thumbprint strings   The thumbprint of the certificate(s) to add to the store(s).
```

### SEE ALSO

* [kfutil stores inventory](kfutil_stores_inventory.md)	 - Commands related to certificate store inventory management

###### Auto generated by spf13/cobra on 1-Dec-2022
