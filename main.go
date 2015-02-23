package main

import (
	"fmt"
	"os"
)

//
func main() {
	handle(ClusterStart())
	handle(StatusStart())
	handle(DecisionStart())
	handle(ActionStart())
	// do some sleep thing
}

//
func handle(err error) {
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
}
