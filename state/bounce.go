package state

import "time"

type (
	Bouncer struct {
		bounce   remoteState
		location string
	}

	BounceString struct {
		Address string
		Method  string
		Timeout time.Duration
		In      string
	}

	BounceBool struct {
		Address string
		Method  string
		Timeout time.Duration
		In      bool
	}

	BounceNil struct {
		Address string
		Method  string
		Timeout time.Duration
		In      Nil
	}
)

func (c remoteState) Bounce(location string) State {
	bounce := Bouncer{
		bounce:   c,
		location: location,
	}
	return bounce
}

func (wrap *StateRPC) BounceString(bounce BounceString, reply *string) error {
	// we need an extra step just incase the next variable gets modified while being used
	// to encode the reply in a Timeout condition
	var next string
	err := call("tcp", bounce.Address, bounce.Timeout, bounce.Method, bounce.In, &next)
	if err == Timeout {
		*reply = "dead"
		return nil
	}
	*reply = next
	return err
}

func (wrap *StateRPC) BounceBool(bounce BounceBool, reply *bool) error {
	return call("tcp", bounce.Address, bounce.Timeout, bounce.Method, bounce.In, reply)
}

func (wrap *StateRPC) BounceNil(bounce BounceNil, reply *Nil) error {
	return call("tcp", bounce.Address, bounce.Timeout, bounce.Method, bounce.In, reply)
}

func (bounce Bouncer) Bounce(location string) State {
	// this should really return an error
	return nil
}

func (bounce Bouncer) Ready() {
	next := BounceNil{
		In:      Nil{},
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.Ready",
	}

	for bounce.bounce.call("StateRPC.BounceNil", next, &Nil{}) != nil {
		<-time.After(time.Second)
	}
}

func (bounce Bouncer) SetSynced(synced bool) error {
	next := BounceBool{
		In:      synced,
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.SetSynced",
	}
	var out bool
	return bounce.bounce.call("StateRPC.BounceBool", next, &out)
}

func (bounce Bouncer) HasSynced() (bool, error) {
	var synced bool
	next := BounceBool{
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.HasSynced",
	}
	err := bounce.bounce.call("StateRPC.BounceBool", next, &synced)
	return synced, err
}

func (bounce Bouncer) Location() string {
	return bounce.location
}

func (bounce Bouncer) GetDataDir() (string, error) {
	var dataDir string
	next := BounceString{
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.GetDataDir",
	}
	err := bounce.bounce.call("StateRPC.BounceString", next, &dataDir)
	return dataDir, err
}

func (bounce Bouncer) GetRole() (string, error) {
	var role string
	next := BounceString{
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.GetRole",
	}
	err := bounce.bounce.call("StateRPC.BounceString", next, &role)
	return role, err
}

func (bounce Bouncer) GetDBRole() (string, error) {
	var dbRole string
	next := BounceString{
		Address: bounce.location,
		Timeout: bounce.bounce.timeout,
		Method:  "StateRPC.GetDBRole",
	}
	err := bounce.bounce.call("StateRPC.BounceString", next, &dbRole)
	return dbRole, err
}

func (bounce Bouncer) SetDBRole(role string) error {
	return NotSupported
}
