// status_test.go tests the functionaly of status.go

package main

import (
	// "fmt"
	// "net"
	// "net/rpc"
	// "strconv"
	"testing"
	"time"

	// "github.com/nanobox-core/scribble"
)

//
// func testStatusStart() error { }

//
func testSetDBRole(t *testing.T) {

	status := &Status{CRole: "primary", DBRole: "initialized", State: "booting", UpdatedAt: time.Now()}
	targetRole := "master"

	//
	status.SetDBRole(targetRole)

	if status.DBRole != targetRole {
		t.Error(
			"For", status.DBRole,
			"expected", targetRole,
			"got", status.DBRole,
		)
	}
}

//
func (s *Status) testSetState(t *testing.T) {

	status := &Status{CRole: "primary", DBRole: "initialized", State: "booting", UpdatedAt: time.Now()}
	targetState := "running"

	//
	status.SetState(targetState)

	if status.DBRole != targetState {
		t.Error(
			"For", status.DBRole,
			"expected", targetState,
			"got", status.DBRole,
		)
	}
}

//
// func testWhoami() (*Status, error) {
// }

//
// func testWhois(role string) (*Status, error) {
// }

//
// func testCluster() ([]*Status, error) {
// }

//
// func (s *Status) testPing(role string, status *Status) error {
// }

//
// func (s *Status) testDemote(source string, status *Status) error {
// }

//
// func testGet(role string, status *Status) error {
// }

//
// func testSave(status *Status) error {
// }
