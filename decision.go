package main

import (
	"time"
)

var lastKnownCluster *[]Status

func DecisionStart() error {
	waitForClusterFull()
	go func() {
		self := WhoAmI()
		if self.CRole == "monitor" {
			fmt.Println("im a monitor.. i dont make decisions")
			return nil
		}
		// figure out what to start as.
		startDb()
		lastKnownCluster = Cluster()
		for {
			if clusterChanges() {
				performAction()
			}
			// check for node dropping
			// if node drops react accordingly
		}
		
	}()
	return nil
}

// this will ping the cluster until it has the
// appropriate number of members
func waitForClusterFull() {
	for {
		if len(Cluster()) > 2 {
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
  	action <- startType("master")
  case "secondary":
  	action <- startType("slave")
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
	case "master":
		// check the other node and see if it is single
		// if not i stay master 
		// if so i go secondary
		other, err := WhoIs(otherRole(self))
		if other.DBRole == "single" {
			return "slave"
		}
		retun "master"
	case "slave":
		return "slave"
	}
	panic("i should have caught all scenarios")
	return def
}

func clusterChanges() bool {
	for _, member, := lastKnownCluster {
		if member != WhoIs(member.CRole) {
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

		return "master"
	case "master":
		// check the other node and see if it is single
		// if not i stay master 
		// if so i go secondary
		if self.CRole == "primary" {
			sec := WhoIs("secondary")
			if sec.DBRole == "single" {
				return "slave"
			}
		}
		retun "master"
	case "slave":
		return "slave"
	}
}

func otherRole(st Status) string {
	if st.CRole == "primary" {
		return "secondary"
	} else {
		return "primary"
	}
	return ""
}