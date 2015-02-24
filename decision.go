package main

import (
	"time"
)

var lastKnownCluster []*Status

//
func DecisionStart() error {
	log.Info("[decision] starting")
	// wait for the cluster to come online
	waitForClusterFull()
	// start the database and perform actions on that database
	go func() {
		self := myself()
		log.Info("[decision] myself %+v", self)
		if self.CRole == "monitor" {
			log.Info("[decision] im a monitor.. i dont make decisions")
			return
		}
		// start the database up
		startDB()
		lastKnownCluster, _ = Cluster()
		// start a timer that will trigger a cluster check
		timer := make(chan bool)
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

			case result := <-advice:
				if result == "demote" && self.DBRole == "master" {
					updateStatusRole("dead(master)")
					actions <- "kill"
				} else {
					log.Info("got a result i am not doing anything with:" +result)
					// what do i do with other advice?
					// if clusterChanges() {
					// 	performAction()
					// }
				}
			case <-timer:
				log.Info("[decision] timer ran out: checking cluster")
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
		c, _ := Cluster()
		if len(c) == 3 {
			log.Info("[decision] members are all online!")
			return
		}
		
		log.Info("[decision] waiting for members (cluster(%d), list(%d))\n",len(c), len(list.Members()))
		time.Sleep(time.Second)
	}
}

// figure out what to start as.
func startDB() {
	log.Info("[decision] Starting Db")
	self := myself()
	switch self.CRole {
	case "primary":
		r := startType("master")
		updateStatusRole(r)
		log.Info("[decision] I am starting as "+r)
		actions <- r
	case "secondary":
		r := startType("slave")
		updateStatusRole(r)
		log.Info("[decision] I am starting as "+r)
		actions <- r
	default:
		log.Info("[decision] Monitors dont do anything.")
	}
}

//
func startType(def string) string {
	self := myself()
	log.Info("[decision] startType: self: %+v", self)
	switch self.DBRole {
	case "initialized":
		return def
	case "single":
		return "master"
	case "master", "dead(master)":
		// check the other node and see if it is single
		// if not i stay master
		// if so i go secondary
		other, _ := Whois(otherRole(self))
		log.Info("[decision] startType: other: %+v", other)
		// if the other guy has transitioned to single
		if other.DBRole == "single" {
			return "slave"
		}
		// if the other guy detected i came back online and is already
		// switching to master
		if other.DBRole == "master" && other.UpdatedAt.After(self.UpdatedAt) {
			return "slave"
		}
		return "master"
	case "slave", "dead(slave)":
		return "slave"
	}
	log.Error("[decision] %+v\n",self)
	panic("i should have caught all scenarios")
	return def
}

//
func clusterChanges() bool {
	c, _ := Cluster()
	if len(lastKnownCluster) != len(c) {
		lastKnownCluster, _ = Cluster()
		return true
	}
	for _, member := range lastKnownCluster {
		other, _ := Whois(member.CRole)
		if member.DBRole != other.DBRole {
			lastKnownCluster, _ = Cluster()
			return true
		}
	}
	return false
}

//
func performAction() {
	self := myself()
	other, _ := Whois(otherRole(self))

	log.Info("[decision] performAction: self: %+v, other: %+v", self, other)
	switch self.DBRole {
	case "single":
		performActionFromSingle(self, other)
	case "master":
		performActionFromMaster(self, other)
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
		log.Info("[decision] performActionFromSingle: going master")
		actions <- "master"
	}
}

//
func performActionFromMaster(self, other *Status) {
	if other != nil && other.DBRole == "slave" {
		// i lost the monitor
		// shouldnt hurt anything
		log.Info("[decision] performActionFromMaster: other is slave .. im doing nothing")
		return
	}
	if other != nil && other.DBRole == "dead(slave)" {
		// my slave has died and i need to transition into single mode
		updateStatusRole("single")
		log.Info("[decision] performActionFromMaster: other is dead .. going single")
		actions <- "single"
	}
	// see if im the odd man out or if it is the other guy
	time.Sleep(10 * time.Second)
	mon, _ := Whois("monitor")
	if mon != nil {
		// the other member died but i can still talk to the monitor
		// i can safely become a single
		updateStatusRole("single")
		log.Info("[decision] performActionFromMaster: other dead but im still in going single")
		actions <- "single"
	} else {
		// I have lost communication with everything else
		// kill my server
		updateStatusRole("dead(master)")
		log.Info("[decision] performActionFromMaster: i am out of cluster SUPUKU")
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
	mon, _ := Whois("monitor")
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
	c, _ := Cluster()
	if other != nil && len(c) == 3 {
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
	status.SetDBRole(r)
	lastKnownCluster, _ = Cluster()
}

//
func otherRole(st *Status) string {
	if st.CRole == "primary" {
		return "secondary"
	}
	return "primary"
}

func myself() *Status {
	for i := 0; i < 10; i++ {
		self, err := Whoami()
		if err == nil {
			return self
		}
		log.Error("Decision: Myself: "+ err.Error())
	}
	panic("Decision: Myself: I never found myself!")
	return nil
}
