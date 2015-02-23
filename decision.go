package main

import (
	"time"
)

func DecisionStart() error {
	waitForClusterFull()
	go func() {
		for {
			monitor, primary, secondary := assignMembers()
		}
		
		case
		Status.State
	}()
	return nil
}

func waitForClusterFull() {
	for {
		if len(Cluster()) > 2 {
			fmt.Println("members are all online!")
			return
		}
		time.Sleep(time.Second)
	}
}

func assignMembers() monitor, primary, secondary Status {
	
}