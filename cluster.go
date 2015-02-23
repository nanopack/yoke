package main

import (
	"github.com/hashicorp/memberlist"
)

var list *memberlist.Memberlist

//
func ClusterStart() error {
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventHandler{true}
	config.Conflict = EventHandler{true}
	config.BindPort = conf.ClusterPort
	config.AdvertisePort = conf.ClusterPort
	l, err := memberlist.Create(config)
	if err != nil {
		log.Error("cluster Error: %s", err.Error())
	}
	list = l
	
	_, err = list.Join(conf.Peers)
	if err != nil {
		log.Error("cluster Error: %s", err.Error())
	}

	return nil
}
