package main

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	// "os"
)

type EventThing struct {
	Count int
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (e EventThing) NotifyJoin(n *memberlist.Node) {
	fmt.Printf("NotifyJoin:%+v", n)
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (e EventThing) NotifyLeave(n *memberlist.Node) {
	fmt.Printf("NotifyLeave:%+v", n)
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (e EventThing) NotifyUpdate(n *memberlist.Node) {
	fmt.Printf("NotifyUpdate:%+v", n)
}

func (e EventThing) NotifyConflict(existing, other *memberlist.Node) {
	// if other == list.LocalNode() {
	// 	fmt.Println("I cannot join that cluster", existing, other)
	// 	os.Exit(1)
	// }
	fmt.Printf("someone tried joing us:\n %+v\n\n%+v", existing, other)
}
