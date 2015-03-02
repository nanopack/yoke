package main

import "github.com/hashicorp/memberlist"

//
var list *memberlist.Memberlist

// ClusterStart
func ClusterStart() error {
	log.Info("[cluster.ClusterStart]")

	// configur memberlist
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventHandler{true}
	config.Conflict = EventHandler{true}
	config.BindPort = conf.AdvertisePort
	config.AdvertiseAddr = conf.AdvertiseIp
	config.AdvertisePort = conf.AdvertisePort
	config.IndirectChecks = 1 // we only have 3 servers; if 1 goes offline we cant check more than 1

	var err error

	// Create the initial memberlist from a safe configuration.
	list, err = memberlist.Create(config)
	if err != nil {
		log.Error("[cluster.ClusterStart] Failed to create memberlist! \n%s\n", err)
	}

	// Join an existing cluster by specifying at least one known member.
	if _, err := list.Join(conf.Peers); err != nil {
		log.Error("[cluster.ClusterStart] Failed to join cluster! \n%s\n", err)
	}

	return nil
}
