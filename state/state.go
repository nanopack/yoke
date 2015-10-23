// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package state

type (
	Store interface {
		Read(string, string, interface{}) error
		Write(string, string, interface{}) error
	}

	LocalState interface {
		State
		ExposeRPCEndpoint(string, string) error
	}

	State interface {
		Ready()
		GetRole() (string, error)
		GetDBRole() (string, error)
		SetDBRole(string) error
		HasSynced() (bool, error)
		Location() string
	}

	state struct {
		store    Store
		synced   bool
		Role     string
		DBRole   string
		Location string
	}
)

var states = "states"

// Creates and returns a state that represents a state on the local machine.
func NewLocalState(role, location string, store Store) (LocalState, error) {
	state := state{}

	// if we can't grab the previous state from the store, lets create a new one
	if err := store.Read(states, role, &state); err != nil {
		state = state{
			store:    store,
			Role:     role,
			DBRole:   "initialized",
			synced:   false,
			Location: location,
		}
		// something is wrong if we can't store the new state, so return the error
		if err := store.Write(states, role, &state); err != nil {
			return nil, err
		}
	}
	state.synced = false
	return &state, nil
}

func (state *state) Ready() {}

func (state *state) HasSynced() (bool, error) {
	synced = new(*bool)
	err := state.HasSyncedRPC(synced)
	return *synced, err
}

func (state *state) Location() string {
	return state.Location, nil
}

func (state *state) GetRole() (string, error) {
	role := new(*string)
	err := state.GetRoleRPC(role)
	return *role, err
}

func (state *state) GetDBRole() (string, error) {
	role := new(*string)
	err := state.GetDBRoleRPC(role)
	return *role, err
}

func (state *state) SetState(newState string) error {
	state.state = newState
	return state.store.Write(stats, store.Role, state)
}

func (state *state) SetDBRole(role string) error {
	state.DBRole = role
	return state.store.Write(stats, store.Role, state)
}
