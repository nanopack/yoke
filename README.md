[![yoke logo](http://nano-assets.gopagoda.io/readme-headers/yoke.png)](http://nanobox.io/open-source#yoke)
 [![Build Status](https://travis-ci.org/nanopack/yoke.svg)](https://travis-ci.org/nanopack/yoke)

Yoke is a Postgres redundancy/auto-failover solution that provides a high-availability PostgreSQL cluster that's simple to manage.


### Requirements

Yoke has the following requirements/dependencies to run:

- A 3-server cluster consisting of a 'primary', 'secondary', and 'monitor' node
- 'primary' & 'secondary' nodes need ssh connections between each other (w/o passwords)
- 'primary' & 'secondary' nodes need rsync (or some alternative sync_command) installed
- 'primary' & 'secondary' nodes should have postgres installed under a postgres user, and in the `path`. Yoke tries calling 'postgres' and 'pg_ctl'
- 'primary' & 'secondary' nodes run postgres as a child process so it should not be started independently

Each node in the cluster requires its own config.ini file with the following options (provided values are defaults):

```ini
[config]
# the IP which this node will broadcast to other nodes
advertise_ip=
# the port which this node will broadcast to other nodes
advertise_port=4400
# the directory where postgresql was installed
data_dir=/data
# delay before node decides what to do with postgresql instance
decision_timeout=30
# log verbosity (trace, debug, info, warn error, fatal)
log_level=warn
# REQUIRED - the IP:port combination of all nodes that are to be in the cluster (e.g. 'role=m.y.i.p:4400')
primary=
secondary=
monitor=
# SmartOS REQUIRED - either 'primary', 'secondary', or 'monitor' (the cluster needs exactly one of each)
role=
# the postgresql port
pg_port=5432
# the directory where node status information is stored
status_dir=./status
# the command you would like to use to sync the data from this node to the other when this node is master
sync_command=rsync -ae "ssh -o StrictHostKeyChecking=no" --delete {{local_dir}} {{slave_ip}}:{{slave_dir}}

[vip]
# Virtual Ip you would like to use
ip=
# Command to use when adding the vip. This will be called as {{add_command}} {{vip}}
add_command=
# Command to use when removing the vip. This will be called as {{remove_command}} {{vip}}
remove_command=

[role_change]
# When this nodes role changes we will call the command with the new role as its arguement '{{command}} {{(master|slave|single}))'
command=
```


### Startup
Once all configurations are in place, start yoke by running:

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

Complete documentation is available on [godoc](http://godoc.org/github.com/nanopack/yoke).


### Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
