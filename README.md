## Yoke

Yoke is a postgres redundancy/auto-failover solution


### Usage

Yoke has the following requirements/dependancies to run:

- a 3 server cluster consisting of a 'primary', 'secondary', and 'monitor' node
- 'primary' & 'secondary' nodes need ssh connections between each other (w/o passwords)
- 'primary' & 'secondary' nodes need rsync (or some alternative sync_command) installed
- 'primary' & 'secondary' nodes should have postgres installed under a postgres user, and in the `path`

Each node in the cluster requires it's own config.ini file with the following options (provided values are defaults):
```
    [config]
    advertise_ip=         # REQUIRED - the IP which this node will broadcast to other nodes
    advertise_port=4400   # the port which this node will broadcast to other nodes
    data_dir=/data/       # the directory where postgresql was installed
    decision_timeout=10   # delay before node dicides what to do with postgresql instance
    log_level=info        # log verbosity (trace, debug, info, warn error, fatal)
    peers=                # REQUIRED - the (comma delimited) IP:port combination of all nodes that are to be in the cluster
    pg_port=5432          # the postgresql port
    role=monitor          # REQUIRED - either 'primary', 'secondary', or 'monitor' (the cluster needs exactly one of each)
    status_dir=./status/  # the directory where node status information is stored
    sync_command="rsync -a --delete {{local_dir}} {{slave_ip}}:{{slave_dir}}" # the command you would like to use to sync the data from this node to the other when this node is master

    [vip]
    ip="1.2.3.4"         # Virtual Ip you would like to use
    add_command          # Command to use when adding the vip. This will be called as {{add_command}} {{vip}}
    remove_command       # Command to use when removeing the vip. This will be called as {{remove_command}} {{vip}}

    [role_change]
    command        # When this nodes role changes we will call the command with the new role as its arguement '{{command}} {{(master|slave|single}))'
```

#### Startup
Once all configurations are in place, to start yoke simply run:

    ./yoke ./primary.ini

NOTE: The file can be named anything, and reside anywhere, all yoke needs is the /path/to/config.ini on startup


### Documentation

Complete documentation is available on [godoc](http://godoc.org/github.com/pagodabox-tools/yoke).


### Contributing
