package main

import "github.com/hashicorp/memberlist"
import "time"
import "os"

//
var list *memberlist.Memberlist

type ClusterWriter struct {
	// just need a couple methods
}

// Write is just a fulfillment of the io.Writer interface
func (c ClusterWriter) Write(d []byte) (int, error) {
	log.Trace("%s %s", "[memberlist]", d)
	return len(d), nil
}

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
	config.LogOutput = ClusterWriter{}
	var err error

	// Create the initial memberlist from a safe configuration.
	list, err = memberlist.Create(config)
	if err != nil {
		log.Error("[cluster.ClusterStart] Failed to create memberlist! \n%s\n", err)
	}

	// Join an existing cluster by specifying at least one known member.
	joinSuccess := false
	for i := 0; i < 10; i++ {
		if _, err := list.Join(conf.Peers); err == nil {
			joinSuccess = true
			break
		} else {
			log.Error("[cluster.ClusterStart] Failed to join cluster! \n%s\n", err)
			log.Debug("retrying %d more times", (9 - i))
			time.Sleep(time.Second)
		}
	}
	if !joinSuccess {
		log.Fatal("I could not successfully join the cluster")
		log.Close()
		os.Exit(1)
	}

	return nil
}
