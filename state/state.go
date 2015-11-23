//
package state

import (
	"io"
)

type (
	Store interface {
		Read(string, string, interface{}) error
		Write(string, string, interface{}) error
	}

	LocalState interface {
		State
		ExposeRPCEndpoint(string, string) (io.Closer, error)
	}

	State interface {
		Ready()
		GetDataDir() (string, error)
		GetRole() (string, error)
		GetDBRole() (string, error)
		SetDBRole(string) error
		HasSynced() (bool, error)
		SetSynced(bool) error
		Location() string
		Bounce(location string) State
	}

	state struct {
		store   Store
		synced  bool
		Role    string
		DBRole  string
		Address string
		DataDir string
	}
)

var states = "states"

// Creates and returns a state that represents a state on the local machine.
func NewLocalState(role, location, dataDir string, store Store) (LocalState, error) {
	newState := state{}

	// if we can't grab the previous state from the store, lets create a new one
	if err := store.Read(states, role, &newState); err != nil {
		newState = state{
			DataDir: dataDir,
			Role:    role,
			DBRole:  "initialized",
			synced:  false,
			Address: location,
		}
		// something is wrong if we can't newState the new state, so return the error
		if err = store.Write(states, role, &newState); err != nil {
			return nil, err
		}
	}
	newState.store = store
	newState.synced = false
	return &newState, nil
}

func (state *state) Ready() {}

func (state *state) Bounce(location string) State {
	// this should really return an error of some sort...
	return nil
}

func (state *state) HasSynced() (bool, error) {
	return state.synced, nil
}

func (state *state) SetSynced(synced bool) error {
	state.synced = synced
	return nil
}

func (state *state) Location() string {
	return state.Address
}

func (state *state) GetDataDir() (string, error) {
	return state.DataDir, nil
}

func (state *state) GetRole() (string, error) {
	return state.Role, nil
}

func (state *state) GetDBRole() (string, error) {
	return state.DBRole, nil
}

func (state *state) SetDBRole(role string) error {
	state.DBRole = role
	return state.store.Write(states, state.Role, state)
}
