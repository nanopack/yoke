package main

import (
	"time"
	"fmt"
)

//
func ActionStart() error {
	for {
		select {
		case act := <-actions:
			doAction(act)
		}
	}

	return nil
}

//
func doAction(act string) {
	switch act {
	case "kill":
		status.SetState("shutting_down")
		time.Sleep(time.Second)
		status.SetState("dieing")
		time.Sleep(time.Second)
		status.SetState("down")
	case "master":
		status.SetState("configuring")
		time.Sleep(time.Second)
		status.SetState("starting/restarting")
		time.Sleep(time.Second)
		status.SetState("running")
	case "slave":
		status.SetState("configuring")
		time.Sleep(time.Second)
		status.SetState("starting/restarting")
		time.Sleep(time.Second)
		status.SetState("running")
	case "single":
		status.SetState("configuring")
		time.Sleep(time.Second)
		status.SetState("starting/restarting")
		time.Sleep(time.Second)
		status.SetState("running")
	default:
		fmt.Println("i dont know what to do with this action: " + act)
	}
}
