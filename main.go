package main

import (
	"fmt"
	"os"
)

//
func main() {
	fmt.Println("0")
	handle(ClusterStart())
	fmt.Println("1")
	handle(StatusStart())
	fmt.Println("2")
	handle(DecisionStart())
	fmt.Println("3")
	handle(ActionStart())
	fmt.Println("4")
	// do some sleep thing
}

//
func handle(err error) {
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
}
