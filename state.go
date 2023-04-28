// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

var (
	Ignored     StateKind = []byte{'I'}
	Unmonitored StateKind = []byte{'U'}
	Monitored   StateKind = []byte{'M'}
	Renamed     StateKind = []byte{'R'}
)

type StateKind []byte

func (me *StateKind) IsIgnored() bool {
	return len(*me) == 1 && (*me)[0] == 'I'
}

func (me *StateKind) IsUnmonitored() bool {
	return len(*me) == 1 && (*me)[0] == 'U'
}

func (me *StateKind) IsMonitored() bool {
	return len(*me) == 1 && (*me)[0] == 'M'
}

func (me *StateKind) IsRenamed() bool {
	return len(*me) == 1 && (*me)[0] == 'R'
}

type StateData struct {
	filename string
	state    StateKind
}

func newState(filename string, state StateKind) *StateData {
	return &StateData{filename: filename, state: state}
}

func newStateFromRaw(filename []byte, state []byte) *StateData {
	return newState(string(filename), StateKind(state))
}

func (me *StateData) Filename() string {
	return me.filename
}

func (me *StateData) IsIgnored() bool {
	return me.state.IsIgnored()
}

func (me *StateData) IsUnmonitored() bool {
	return me.state.IsUnmonitored()
}

func (me *StateData) IsMonitored() bool {
	return me.state.IsMonitored()
}

func (me *StateData) IsRenamed() bool {
	return me.state.IsRenamed()
}
