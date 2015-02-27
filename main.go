package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//
func main() {
	handle(ClusterStart())
	handle(StatusStart())
	handle(DecisionStart())
	handle(ActionStart())
	// signal Handle
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, os.Kill, syscall.SIGQUIT, syscall.SIGHUP)

	// Block until a signal is received.
	for {
		s := <-c
		switch s {
		case syscall.SIGINT, os.Kill, syscall.SIGQUIT:
			// kill the database then quit
			log.Info("Signal Recieved: %s", s.String())
			if conf.Role == "monitor" {
				log.Info("shutting down")
			} else {
				log.Info("Killing Database")
				actions <- "kill"
				// called twice because the first call returns when the job is picked up
				// the second call returns when the first job is complete
				actions <- "kill"
			}
			log.Close()
			os.Exit(0)
		case syscall.SIGHUP:
			// demote
			log.Info("Signal Recieved: %s", s.String())
			log.Info("advising a demotion")
			advice <- "demote"
		}
	}

}

//
func handle(err error) {
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
}
