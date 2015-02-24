package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/memberlist"
)

type EventHandler struct {
	Active bool
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (e EventHandler) NotifyJoin(n *memberlist.Node) {
	fmt.Printf("NotifyJoin:%+v", n)
	go func() {
		advice <- "join" + n.Name
	}()
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (e EventHandler) NotifyLeave(n *memberlist.Node) {
	fmt.Printf("NotifyLeave:%+v", n)
	go func() {
		advice <- "leave" + n.Name
	}()

}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (e EventHandler) NotifyUpdate(n *memberlist.Node) {
	fmt.Printf("NotifyUpdate:%+v", n)
	go func() {
		advice <- "update" + n.Name
	}()

}

//
func (e EventHandler) NotifyConflict(existing, other *memberlist.Node) {
	defer func() {
		recover()
		fmt.Println("I cannot join that cluster")
		os.Exit(1)
	}()
	fmt.Println(len(list.Members()))
	fmt.Printf("someone tried joing us:\n %+v\n\n%+v", existing, other)
}
