package main

import "fmt"
import "os"
import "errors"

var list *memberlist.Memberlist

func main() {
	err := ClusterStart()
	handle(err)
	err := RpcStart()
	handle(err)
	err := DecisionStart()
	handle(err)
	err := ActionStart()
	handle(err)
	// hang
}

func handle(err error) {
	if err != nil {
		fmt.Println("error: " + err)
		os.Exit(1)
	}
}
