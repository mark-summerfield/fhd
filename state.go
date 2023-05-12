// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "fmt"

type StateVal struct {
	Sid       SID // Most recent SID the corresponding file was saved into
	Monitored bool
	Renamed   bool
	FileKind  fileKind
}

func newStateVal(sid SID, monitored bool, fileKind fileKind) StateVal {
	return StateVal{Sid: sid, Monitored: monitored, FileKind: fileKind}
}

func (me StateVal) String() string {
	monitored := "M"
	if !me.Monitored {
		monitored = "U"
	}
	renamed := "r"
	if !me.Renamed {
		renamed = " "
	}
	return fmt.Sprintf("%s%s#%d:%s", monitored, renamed, me.Sid, me.FileKind)
}

func (me StateVal) marshal() []byte {
	raw := make([]byte, 0, 10)
	raw = append(raw, me.Sid.marshal()...)
	var monitored byte = 'M'
	if !me.Monitored {
		monitored = 'U'
	}
	raw = append(raw, monitored)
	var renamed byte = 'r'
	if !me.Renamed {
		renamed = ' '
	}
	raw = append(raw, renamed)
	return append(raw, byte(me.FileKind))
}

func unmarshalStateVal(raw []byte) StateVal {
	var stateVal StateVal
	index := sidSize
	stateVal.Sid = unmarshalSid(raw[:index])
	stateVal.Monitored = raw[index] == 'M'
	index++
	stateVal.Renamed = raw[index] == 'r'
	if len(raw) > sidSize {
		index++
		stateVal.FileKind = fileKind(raw[index])
	}
	return stateVal
}

type StateItem struct {
	Filename string
	StateVal
}

func newState(filename string, stateVal StateVal) *StateItem {
	return &StateItem{Filename: filename, StateVal: stateVal}
}

func newStateFromRaw(rawFilename []byte, rawStateVal []byte) *StateItem {
	return newState(string(rawFilename), unmarshalStateVal(rawStateVal))
}

func (me StateItem) String() string {
	return fmt.Sprintf("%q%s", me.Filename, me.StateVal)
}
