// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

// status_test.go tests the functionaly of status.go

package main

import (
	// "fmt"
	// "net"
	// "net/rpc"
	// "strconv"
	"testing"
	"time"
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
// func testWhoami() {
// }

//
// func testWhois() {
// }

//
// func testCluster() {
// }

//
// func (s *Status) testPing() {
// }

//
// func (s *Status) testDemote() {
// }

//
// func testGet() {
// }

//
// func testSave() {
// }
