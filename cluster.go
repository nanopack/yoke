package main

import (
	"github.com/hashicorp/memberlist"
)

var list *memberlist.Memberlist

// ClusterStart
func ClusterStart() error {

	log.Debug("[cluster.ClusterStart] config: %+v", conf)

	//
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventHandler{true}
	config.Conflict = EventHandler{true}
	// config.BindAddr = conf.ClusterIP/"all"?
	config.BindPort = conf.ClusterPort
	config.AdvertiseAddr = conf.ClusterIP
	config.AdvertisePort = conf.ClusterPort
	config.IndirectChecks = 1

	//
	l, err := memberlist.Create(config)
	if err != nil {
		log.Error("[cluster.ClusterStart] failed to create memberlist!\n%s\n", err)
	}

	list = l

	//
	_, err = list.Join(conf.Peers)
	if err != nil {
		log.Error("[cluster.ClusterStart] failed to join cluster!\n%s\n", err)
	}

	return nil
}
