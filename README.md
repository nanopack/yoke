## Yoke

Yoke is a postgres redundancy/auto-failover solution


### Usage

##### Dependancies
* 3-node cluster consisting of a `primary`, `secondary`, and `monitor` node
* `primary` and `secondary` nodes need postgres installed and in the `path`
* `primary` and `secondary` nodes need ssh connections between each other (w/o passwords)
* `primary` and `secondary` nodes need rsync (or some alternative sync_command) installed

Each node in the cluster requires it's own config file:

path/to/configs/monitor.ini:
[config]
  log_level=debug
  role=monitor
  advertise_ip=192.168.0.4
  advertise_port=4400
  peers=192.168.0.2:4400,192.168.0.3:4400,192.168.0.4:4400
  decision_timeout=10


### Documentation

Complete documentation is available on [godoc](http://godoc.org/github.com/pagodabox-tools/yoke).


### Contributing
