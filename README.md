# DRACLI

A quick and dirty CLI/client library for the Integrated Dell Remote Access Controller v6.

## CLI Usage

I recommend that you setup a new user in the iDRAC admin specifically for this CLI tool. It will prevent you from getting locked out of the main account if you log in too many times.


If you do happen to lock yourself out, you can SSH into the iDRAC with your root credentials and run `racadm racreset`


```sh
# start by using login - stores at ~/.dracli/credentials.json
dracli login -u root -p calvin -h 10.0.0.5

# manage power state
dracli power [on|off|nmi|graceful_shutdown|cold_reboot|warm_reboot]

# query info about your server
dracli query  pwState sysDesc fans

# query info about your server continuously
dracli query -watch 1s pwState sysDesc fans

# list help 
dracli help

# log out removes credentials 
dracli logout 
```


## Future Features

- xml output
- manage/query multiple servers simultaneously
- remote console
  - run virtual console locally (requires java)
- manage user accounts


## LICENSE 

MIT