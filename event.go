// event.go

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/memberlist"
)

// EventHandler is a implentation of memberlists Deligate interface. It will be
// used to alert our decision system when node activity happens.
type EventHandler struct {
	Active bool //
}

// NotifyJoin is invoked when a node is detected to have joined. The Node argument
// must not be modified. We send advice to the decision engine and let it decide
// what to do
func (e EventHandler) NotifyJoin(n *memberlist.Node) {
	log.Debug("[event.NotifyJoin] '%s' joined the cluster...\n", n.Name)
	go func() {
		advice <- "join" + n.Name
	}()
}

// NotifyLeave is invoked when a node is detected to have left. The Node argument
// must not be modified. We send advice to the decision engine and let it decide
// what to do
func (e EventHandler) NotifyLeave(n *memberlist.Node) {
	log.Debug("[event.NotifyLeave] '%s' left the cluster...\n", n.Name)
	go func() {
		advice <- "leave" + n.Name
	}()
}

// NotifyUpdate is invoked when a node is detected to have updated, usually
// involving the meta data. The Node argument must not be modified. We send advice
// to the decision engine and let it decide what to do
func (e EventHandler) NotifyUpdate(n *memberlist.Node) {
	log.Debug("[event.NotifyUpdate] '%s' was updated\n", n.Name)
	go func() {
		advice <- "update" + n.Name
	}()
}

// NotifyConflict detects if we colide with a member of the same name and are
// unable to join
func (e EventHandler) NotifyConflict(existing, other *memberlist.Node) {

	//
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("[event.NotifyConflict] '%s' already exists in cluster... unable to join! Exiting...\n", existing.Name)
			log.Close()
			os.Exit(1)
<<<<<<< Updated upstream
		}
=======
    }

<<<<<<< Updated upstream
<<<<<<< Updated upstream
>>>>>>> Stashed changes
=======
>>>>>>> Stashed changes
=======
>>>>>>> Stashed changes
	}()

	fmt.Println(len(list.Members()))
}
