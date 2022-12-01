## kfutil login

User interactive login to Keyfactor. Stores the credentials in the config file '$HOME/.keyfactor/command_config.json'.

### Synopsis

Will prompt the user for a username and password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for username and password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables: KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, 
KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN.

WARNING: The username and password will be stored in the config file in plain text at: 
'$HOME/.keyfactor/command_config.json.'


```
kfutil login [flags]
```

### Options

```
  -c, --config string   config file (default is $HOME/.keyfactor/%s)
  -h, --help            help for login
      --no-prompt       Do not prompt for username and password
```

### SEE ALSO

* [kfutil](kfutil.md)	 - Keyfactor CLI utilities

###### Auto generated by spf13/cobra on 1-Dec-2022