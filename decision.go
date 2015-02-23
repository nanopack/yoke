package main

import (
	"fmt"
	"time"
)

var lastKnownCluster *[]Status

func DecisionStart() error {
	waitForClusterFull()
	go func() {
		self := WhoAmI()
		if self.CRole == "monitor" {
			fmt.Println("im a monitor.. i dont make decisions")
			return
		}
		timer = make(chan float64)
		go func() {
			for {
				time.Sleep(time.Second * 10)
				timer <- float64(t)
			}
		}()

		// figure out what to start as.
		startDb()
		lastKnownCluster = Cluster()
		for {
			select {
			// im no good at advice yet

			//   case result := <- advice :
			//   	fmt.Print(result)
			// if clusterChanges() {
			// 	performAction()
			// }
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
func startDb() {
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

func startType(def string) string {
	self := WhoAmI()
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
		if other.DBRole == "single" {
			return "slave"
		}
		return "master"
	case "slave", "dead(slave)":
		return "slave"
	}
	panic("i should have caught all scenarios")
	return def
}

func clusterChanges() bool {
	for _, member := range lastKnownCluster {
		if member.DBRole != WhoIs(member.CRole).DBRole {
			lastKnownCluster = Cluster()
			return true
		}
	}
	return false
}

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

func performActionFromSingle(self, other Status) {
	if other != nil {
		// i was in single but the other node came back online
		// I should be safe to assume master
		updateStatusRole("master")
		actions <- "master"
	}
}

func performActionFromMaste(self, other Status) {
	if other != nil {
		// i lost the monitor
		// shouldnt hurt anything
		return
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

func performActionFromSlave(self, other Status) {
	if other != nil {
		// i probably lost the monitor
		// shouldnt hurt anything
		return
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

func performActionFromDead(self, other Status) {
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

func updateStatusRole(r string) {
	status.UpdateRole(r)
	lastKnownCluster = Cluster()
}

func otherRole(st Status) string {
	if st.CRole == "primary" {
		return "secondary"
	} else {
		return "primary"
	}
	return ""
}
