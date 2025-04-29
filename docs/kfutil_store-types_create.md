## kfutil store-types create

Create a new certificate store type in Keyfactor.

### Synopsis

Create a new certificate store type in Keyfactor.

```
kfutil store-types create [flags]
```

### Options

```
  -a, --all                Create all store types.
  -f, --from-file string   Path to a JSON file containing certificate store type data for a single store.
  -b, --git-ref string     The git branch or tag to reference when pulling store-types from the internet. (default "main")
  -h, --help               help for create
  -l, --list               List valid store types.
  -n, --name string        Short name of the certificate store type to get. Valid choices are: AKV, AWS-ACM, Akamai, AppGwBin, AzureApp, AzureApp2, AzureAppGw, AzureSP, AzureSP2, BIPCamera, CiscoAsa, CitrixAdc, DataPower, F5-BigIQ, F5-CA-REST, F5-SL-REST, F5-WS-REST, FortiWeb, Fortigate, GCPLoadBal, GcpCertMgr, HCVKV, HCVKVJKS, HCVKVP12, HCVKVPEM, HCVKVPFX, HCVPKI, IISU, Imperva, K8SCert, K8SCluster, K8SJKS, K8SNS, K8SPKCS12, K8SSecret, K8STLSSecr, MOST, Nmap, PaloAlto, RFDER, RFJKS, RFKDB, RFORA, RFPEM, RFPkcs12, SAMPLETYPE, Signum, VMware-NSX, WinCerMgmt, WinCert, WinSql, f5WafCa, f5WafTls, iDRAC
  -r, --repo string        The repository to pull store-types definitions from. (default "kfutil")
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

* [kfutil store-types](kfutil_store-types.md)	 - Keyfactor certificate store types APIs and utilities.

###### Auto generated on 29-Apr-2025
