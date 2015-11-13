[![yoke logo](http://nano-assets.gopagoda.io/readme-headers/yoke.png)](http://nanobox.io/open-source#yoke)
 [![Build Status](https://travis-ci.org/nanopack/yoke.svg)](https://travis-ci.org/nanopack/yoke)

Yoke is a Postgres redundancy/auto-failover solution that provides a high-availability PostgreSQL cluster that's simple to manage.


### Requirements

Yoke has the following requirements/dependancies to run:

- A 3-server cluster consisting of a 'primary', 'secondary', and 'monitor' node
- 'primary' & 'secondary' nodes need ssh connections between each other (w/o passwords)
- 'primary' & 'secondary' nodes need rsync (or some alternative sync_command) installed
- 'primary' & 'secondary' nodes should have postgres installed under a postgres user, and in the `path`. Yoke tries calling 'postgres' and 'pg_ctl'
- 'primary' & 'secondary' nodes run postgres as a child process so it should not be started independently

Each node in the cluster requires it's own config.ini file with the following options (provided values are defaults):

```ini
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
  sync_command='rsync -a --delete {{local_dir}} {{slave_ip}}:{{slave_dir}}' # the command you would like to use to sync the data from this node to the other when this node is master. This uses Mustache style templating so Yoke can fill in the {{local_dir}}, {{slave_ip}}, {{slave_dir}} if you want to use them.

[vip]
  ip="1.2.3.4"          # Virtual Ip you would like to use
  add_command           # Command to use when adding the vip. This will be called as {{add_command}} {{vip}}
  remove_command        # Command to use when removeing the vip. This will be called as {{remove_command}} {{vip}}

[role_change]
  command               # When this nodes role changes we will call the command with the new role as its arguement '{{command}} {{(master|slave|single}))'
```


### Startup
Once all configurations are in place, Start yoke by running:

```
./yoke ./primary.ini
```

**Note:** The ini file can be named anything and reside anywhere. All Yoke needs is the /path/to/config.ini on startup.


### Yoke CLI - yokeadm

Yoke comes with its own CLI, yokeadm, that allows for limited introspection into the cluster.

#### Building the CLI:

```
cd ./yokeadm
go build
./yokeadm
```

##### Usage:

```
yokeadm (<COMMAND>:<ACTION> OR <ALIAS>) [GLOBAL FLAG] <POSITIONAL> [SUB FLAGS]
```

##### Available Commands:

- list   : Returns status information for all nodes in the cluster
- demote : Advises a node to demote

### Documentation

Complete documentation is available on [godoc](http://godoc.org/github.com/nanobox-io/yoke).


### Contributing
[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
