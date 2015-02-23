package main

import (
	"github.com/hashicorp/memberlist"
)

var list *memberlist.Memberlist

func ClusterStart() error {
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventHandler{0}
	config.Conflict = EventHandler{0}
	config.BindPort = conf.ClusterPort
	config.AdvertisePort = conf.ClusterPort
	list, err := memberlist.Create(config)
	if err != nil {
		return err
	}

	_, err = list.Join(conf.Peers)
	if err != nil {
		return err
	}

	return nil
}
