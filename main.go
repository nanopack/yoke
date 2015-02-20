package main

import "fmt"
import "github.com/hashicorp/memberlist"
import "time"

var list *memberlist.Memberlist

func main() {
	config := memberlist.DefaultLANConfig()
	config.Name = conf.Role
	config.Events = EventThing{0}
	config.Conflict = EventThing{0}
	config.BindPort = conf.ClusterPort
	config.AdvertisePort = conf.ClusterPort
	fmt.Printf("%+v\n\n", config)
	list, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	fmt.Printf("%+v\n\n", list)

	n, err := list.Join(conf.Peers)
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}
	fmt.Print(n)
	fmt.Printf("%+v\n\n", config)
	// Ask for members of the cluster
	for _, member := range list.Members() {
		fmt.Printf("Member: %s %d\n", member.Name, member.Port)
	}
	for {
		time.Sleep(1000 * time.Second)
	}

}
