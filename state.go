// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "fmt"

type StateInfo struct {
	Monitored bool
	Sid       SID // Most recent SID the corresponding file was saved into
	MimeType  string
}

func newStateInfo(monitored bool, sid SID, mimeType string) StateInfo {
	return StateInfo{Monitored: monitored, Sid: sid, MimeType: mimeType}
}

func (me StateInfo) String() string {
	monitored := "M"
	if !me.Monitored {
		monitored = "U"
	}
	return fmt.Sprintf("%s#%d:%s", monitored, me.Sid, me.MimeType)
}

func (me StateInfo) Marshal() []byte {
	raw := make([]byte, 0, 9)
	raw = append(raw, me.Sid.Marshal()...)
	var monitored byte = 'M'
	if !me.Monitored {
		monitored = 'U'
	}
	raw = append(raw, monitored)
	return append(raw, []byte(me.MimeType)...)
}

func UnmarshalStateInfo(raw []byte) StateInfo {
	return newStateInfo(raw[8] == 'M', UnmarshalSid(raw[:8]),
		string(raw[9:]))
}

type StateData struct {
	Filename string
	StateInfo
}

func newState(filename string, stateInfo StateInfo) *StateData {
	return &StateData{Filename: filename, StateInfo: stateInfo}
}

func newStateFromRaw(rawFilename []byte, rawStateInfo []byte) *StateData {
	return newState(string(rawFilename), UnmarshalStateInfo(rawStateInfo))
}
