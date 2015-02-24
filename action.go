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
			log.Info("[action] new action: "+act)
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
		time.Sleep(time.Second * 20)
		status.SetState("dieing")
		time.Sleep(time.Second * 20)
		status.SetState("down")
	case "master":
		status.SetState("configuring")
		time.Sleep(time.Second * 20)
		status.SetState("starting/restarting")
		time.Sleep(time.Second * 20)
		status.SetState("running")
	case "slave":
		status.SetState("configuring")
		time.Sleep(time.Second * 20)
		status.SetState("starting/restarting")
		time.Sleep(time.Second * 20)
		status.SetState("running")
	case "single":
		status.SetState("configuring")
		time.Sleep(time.Second * 20)
		status.SetState("starting/restarting")
		time.Sleep(time.Second * 20)
		status.SetState("running")
	default:
		fmt.Println("i dont know what to do with this action: " + act)
	}
}
