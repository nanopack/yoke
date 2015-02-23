package main

import (
	"fmt"
	"time"
)

var lastKnownCluster *[]Status

//
func DecisionStart() error {
	// wait for the cluster to come online
	waitForClusterFull()
	// start the database and perform actions on that database
	go func() {
		self, err := WhoAmI()
		if self.CRole == "monitor" {
			fmt.Println("im a monitor.. i dont make decisions")
			return
		}
		// start the database up
		startDB()
		lastKnownCluster, err = Cluster()
		// start a timer that will trigger a cluster check
		timer = make(chan bool)
		go func() {
			for {
				time.Sleep(time.Second * 10)
				timer <- true
			}
		}()

		// check server on timer and listen for advice
		// if you notice a problem perform an action
		for {
			select {


		  case result := <- advice:
		    fmt.Print(result)
		    if result == "demote" && self.DBRole == "master" {
					updateStatusRole("dead(master)")
					actions <- "kill"
		    } else {
			    // what do i do with other advice?
					// if clusterChanges() {
					// 	performAction()
					// }
		    }
			case <-timer:
				fmt.Println("timer ran out: checking cluster")
				if clusterChanges() {
					performAction()
				}
			}
		}
	}()
	return nil
}

// this will ping the cluster until it has the
// appropriate number of members
func waitForClusterFull() {
	for {
		if len(Cluster()) == 3 {
			fmt.Println("members are all online!")
			return
		}
		time.Sleep(time.Second)
	}
}

// figure out what to start as.
func startDB() {
	self := WhoAmI()
	switch self.CRole {
	case "primary":
		updateStatusRole("master")
		actions <- startType("master")
	case "secondary":
		updateStatusRole("master")
		actions <- startType("slave")
	default:
		fmt.Println("Monitors dont do anything.")
	}
}

//
func startType(def string) string {
	self, err := WhoAmI()
	switch self.DBRole {
	case "initialized":
		return def
	case "single":
		return "master"
	case "master", "dead(master)":
		// check the other node and see if it is single
		// if not i stay master
		// if so i go secondary
		other, err := WhoIs(otherRole(self))
		// if the other guy has transitioned to single
		if other.DBRole == "single" {
			return "slave"
		}
		// if the other guy detected i came back online and is already
		// switching to master
		if other.DBRole == "master" && other.UpdatedAt > self.UpdatedAt {
			return "slave"
		}
		return "master"
	case "slave", "dead(slave)":
		return "slave"
	}
	panic("i should have caught all scenarios")
	return def
}

//
func clusterChanges() bool {
	if len(lastKnownCluster) != len(Cluster()) {
		lastKnownCluster = Cluster()
		return true
	}
	for _, member := range lastKnownCluster {
		if member.DBRole != WhoIs(member.CRole).DBRole {
			lastKnownCluster = Cluster()
			return true
		}
	}
	return false
}

//
func performAction() {
	self, err := WhoAmI()
	other, err := WhoIs(otherRole(self))

	switch self.DBRole {
	case "single":
		performActionFromSingle(self, other)
	case "master":
		performActionFromMaste(self, other)
	case "slave":
		performActionFromSlave(self, other)
	case "dead(master)", "dead(slave)":
		performActionFromDead(self, other)
	}
}

//
func performActionFromSingle(self, other *Status) {
	if other != nil {
		// i was in single but the other node came back online
		// I should be safe to assume master
		updateStatusRole("master")
		actions <- "master"
	}
}

//
func performActionFromMaste(self, other *Status) {
	if other != nil && other.DBRole == "slave" {
		// i lost the monitor
		// shouldnt hurt anything
		return
	}
	if other != nil && other.DBRole == "dead(slave)" {
		// my slave has died and i need to transition into single mode
		updateStatusRole("single")
		actions <- "single"
	}
	// see if im the odd man out or if it is the other guy
	time.Sleep(10 * time.Second)
	mon, err := WhoIs("monitor")
	if mon != nil {
		// the other member died but i can still talk to the monitor
		// i can safely become a single
		updateStatusRole("single")
		actions <- "single"
	} else {
		// I have lost communication with everything else
		// kill my server
		updateStatusRole("dead(master)")
		actions <- "kill"
	}
}

//
func performActionFromSlave(self, other *Status) {
	if other != nil && other.DBRole == "master" {
		// i probably lost the monitor
		// shouldnt hurt anything
		return
	}
	if other != nil && other.DBRole == "dead(master)" {
		// my master has died and i need to transition into single mode
		updateStatusRole("single")
		actions <- "single"
	}

	// see if im the odd man out or if it is the other guy
	time.Sleep(10 * time.Second)
	mon, err := WhoIs("monitor")
	if mon != nil {
		// the other member died but i can still talk to the monitor
		// i can safely become a single
		updateStatusRole("single")
		actions <- "single"
	} else {
		// I have lost communication with everything else
		// kill my server
		updateStatusRole("dead(slave)")
		actions <- "kill"
	}
}

//
func performActionFromDead(self, other *Status) {
	if other != nil && len(Cluster()) == 3 {
		switch self.DBRole {
		case "dead(master)":
			newRole := startType("master")
			updateStatusRole(newRole)
			actions <- newRole
		case "dead(slave)":
			newRole := startType("slave")
			updateStatusRole(newRole)
			actions <- newRole
		default:
			panic("i dont know how to be a " + self.DBRole)
		}
	}
}

//
func updateStatusRole(r string) {
	status.UpdateRole(r)
	lastKnownCluster = Cluster()
}

//
func otherRole(st *Status) string {
	if st.CRole == "primary" {
		return "secondary"
	}
	return "primary"
}
