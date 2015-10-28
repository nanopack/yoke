// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package state_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/nanobox-io/yoke/state"
	"github.com/nanobox-io/yoke/state/mock"
	"testing"
	"time"
)

var err = errors.New("general")

func TestLocal(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	store := mock_state.NewMockStore(ctrl)

	store.EXPECT().Read("states", "something", gomock.Any()).Return(err).Times(2)
	store.EXPECT().Write("states", "something", gomock.Any()).Return(err)

	out, err := state.NewLocalState("something", "wherever", store)
	if err == nil {
		test.Log("should have gotten an error", out, err)
		test.FailNow()
	}

	store.EXPECT().Write("states", "something", gomock.Any()).Return(nil)
	local, err := state.NewLocalState("something", "wherever", store)
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	testState(local, store, test)

	// now for specific tests to local
	store.EXPECT().Write("states", "something", gomock.Any())
	err = local.SetDBRole("testing")
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	dbRole, err := local.GetDBRole()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}
	if dbRole != "testing" {
		test.Log("wrong dbrole was returned")
		test.Fail()
	}

	if local.Location() != "wherever" {
		test.Log("wrong location was returned")
		test.Fail()
	}
}

func TestRpc(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	store := mock_state.NewMockStore(ctrl)

	store.EXPECT().Read("states", "something", gomock.Any()).Return(err)
	store.EXPECT().Write("states", "something", gomock.Any()).Return(nil)
	local, err := state.NewLocalState("something", "wherever", store)
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	err = local.ExposeRPCEndpoint("tcp", "127.0.0.1:1234")
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	client := state.NewRemoteState("tcp", "127.0.0.1:1234", time.Second)

	testState(client, store, test)

	// now for tests specific to remote states

	err = client.SetDBRole("testing")
	if err == nil {
		test.Log("should not have been able to update the db state from remote")
		test.Fail()
	}

	if client.Location() != "127.0.0.1:1234" {
		test.Log("wrong location was returned")
		test.Fail()
	}
}

func testState(client state.State, store *mock_state.MockStore, test *testing.T) {
	role, err := client.GetRole()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}
	if role != "something" {
		test.Log("wrong role was returned")
		test.Fail()
	}

	dbRole, err := client.GetDBRole()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}
	if dbRole != "initialized" {
		test.Log("wrong dbrole was returned", dbRole)
		test.Fail()
	}

	synced, err := client.HasSynced()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	if synced {
		test.Log("it should not have been in sync")
		test.Fail()
	}

	err = client.SetSynced(true)
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	synced, err = client.HasSynced()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	if !synced {
		test.Log("it should have been in sync")
		test.Fail()
	}

	// really doesn't do anything...
	client.Ready()
}
