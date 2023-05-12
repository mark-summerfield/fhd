// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type SaveVal struct {
	When    time.Time
	Comment string
}

type SaveItem struct {
	Sid SID
	SaveVal
}

func newSaveItem(sid SID, when time.Time, comment string) SaveItem {
	return SaveItem{Sid: sid, SaveVal: SaveVal{When: when, Comment: comment}}
}

func newInvalidSaveItem() SaveItem {
	return SaveItem{Sid: InvalidSID}
}

func (me *SaveItem) IsValid() bool {
	return me.Sid.IsValid()
}

func (me *SaveItem) String() string {
	return fmt.Sprintf("%d@%s%q", me.Sid,
		strings.ReplaceAll(me.When.Format(time.DateTime), " ", "T"),
		me.Comment)
}

func (me SaveVal) marshal() ([]byte, error) {
	raw := make([]byte, 0)
	rawWhen, err := me.When.MarshalBinary()
	if err != nil {
		return raw, err
	}
	raw = append(raw, byte(len(rawWhen)))
	raw = append(raw, rawWhen...)
	raw = append(raw, []byte(me.Comment)...)
	return raw, nil
}

func unmarshalSaveVal(raw []byte) (SaveVal, error) {
	var saveVal SaveVal
	if len(raw) == 0 {
		return saveVal, errors.New("can't unmarshal empty saveval")
	}
	index := int(raw[0]) + 1
	when, err := unmarshalTime(raw[1:index])
	if err != nil {
		return saveVal, err
	}
	var comment string
	if len(raw) > index {
		comment = string(raw[index:])
	}
	saveVal.When = when
	saveVal.Comment = comment
	return saveVal, nil
}
