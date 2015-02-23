package main

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"os"
)

type EventThing struct {
	Count int
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (e EventThing) NotifyJoin(n *memberlist.Node) {
	fmt.Printf("NotifyJoin:%+v", n)
	advice <- "join"+n.Name
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (e EventThing) NotifyLeave(n *memberlist.Node) {
	fmt.Printf("NotifyLeave:%+v", n)
	advice <- "leave"+n.Name
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (e EventThing) NotifyUpdate(n *memberlist.Node) {
	fmt.Printf("NotifyUpdate:%+v", n)
	advice <- "update"+n.Name
}

func (e EventThing) NotifyConflict(existing, other *memberlist.Node) {
	defer func() {
		recover()
		fmt.Println("I cannot join that cluster")
		os.Exit(1)
	}()
	fmt.Println(len(list.Members()))
	fmt.Printf("someone tried joing us:\n %+v\n\n%+v", existing, other)
}
