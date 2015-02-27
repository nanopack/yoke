package main

import (
	"github.com/hashicorp/memberlist"
)

var list *memberlist.Memberlist

// ClusterStart
func ClusterStart() error {

	log.Debug("[cluster.ClusterStart] config: %+v", conf)

	// We could make some adjustments here
	// indirectChecks is set to 1 because we only have 3 servers
	// if 1 goes offline we cant check more then 1
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventHandler{true}
	config.Conflict = EventHandler{true}
	config.BindPort = conf.AdvertisePort
	config.AdvertiseAddr = conf.AdvertiseIp
	config.AdvertisePort = conf.AdvertisePort
	config.IndirectChecks = 1

	// create a new member list.
	l, err := memberlist.Create(config)
	if err != nil {
		log.Error("[cluster.ClusterStart] failed to create memberlist!\n%s\n", err)
	}

	list = l

	// join our new member list to an existing one if they exist
	_, err = list.Join(conf.Peers)
	if err != nil {
		log.Error("[cluster.ClusterStart] failed to join cluster!\n%s\n", err)
	}

	return nil
}
