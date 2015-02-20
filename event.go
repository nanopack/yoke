package main

import "fmt"
import "github.com/hashicorp/memberlist"


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
func (e EventThing) NotifyConflict(existing, other *memberlist.Node) {
	fmt.Println("there is a conflict", existing, other)
}

    // NotifyUpdate is invoked when a node is detected to have
    // updated, usually involving the meta data. The Node argument
    // must not be modified.
func (e EventThing) NotifyUpdate(n *memberlist.Node) {
	fmt.Printf("NotifyUpdate:%+v", n)
}
