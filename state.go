// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

var (
	Monitored   StateKind = []byte{'M'}
	Unmonitored StateKind = []byte{'U'}
	Ignored     StateKind = []byte{'I'}
)

type StateKind []byte

func (me *StateKind) String() string {
	return string(*me)
}

func (me *StateKind) Equal(other StateKind) bool {
	return len(*me) == 1 && len(other) == 1 && (*me)[0] == other[0]
}

func (me *StateKind) IsMonitored() bool {
	return len(*me) == 1 && (*me)[0] == 'M'
}

func (me *StateKind) IsUnmonitored() bool {
	return len(*me) == 1 && (*me)[0] == 'U'
}

func (me *StateKind) IsIgnored() bool {
	return len(*me) == 1 && (*me)[0] == 'I'
}

type StateInfo struct {
	State StateKind
	Sid   SID // Most recent SID the corresponding file was saved into
}

func newStateInfo(state StateKind, sid SID) StateInfo {
	return StateInfo{State: state, Sid: sid}
}

func (me StateInfo) Marshal() []byte {
	raw := make([]byte, 0, 9)
	raw = append(raw, MarshalSid(me.Sid)...)
	return append(raw, me.State...)
}

func UnmarshalStateInfo(raw []byte) StateInfo {
	return newStateInfo(StateKind(raw[8:]), UnmarshalSid(raw[:8]))
}

type StateData struct {
	filename string
	StateInfo
}

func newState(filename string, stateInfo StateInfo) *StateData {
	return &StateData{filename: filename, StateInfo: stateInfo}
}

func newStateFromRaw(rawFilename []byte, rawStateInfo []byte) *StateData {
	return newState(string(rawFilename), UnmarshalStateInfo(rawStateInfo))
}

func (me *StateData) Filename() string {
	return me.filename
}

func (me *StateData) IsMonitored() bool {
	return me.State.IsMonitored()
}

func (me *StateData) IsUnmonitored() bool {
	return me.State.IsUnmonitored()
}

func (me *StateData) IsIgnored() bool {
	return me.State.IsIgnored()
}

type FilenameSid struct {
	Filename string
	Sid      SID
}
