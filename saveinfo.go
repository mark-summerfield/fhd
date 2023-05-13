// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type SaveInfoVal struct {
	When    time.Time
	Comment string
}

type SaveInfoItem struct {
	Sid SID
	SaveInfoVal
}

func newSaveInfoItem(sid SID, when time.Time, comment string) SaveInfoItem {
	return SaveInfoItem{Sid: sid,
		SaveInfoVal: SaveInfoVal{When: when, Comment: comment}}
}

func newInvalidSaveInfoItem() SaveInfoItem {
	return SaveInfoItem{Sid: InvalidSID}
}

func (me *SaveInfoItem) IsValid() bool {
	return me.Sid.IsValid()
}

func (me *SaveInfoItem) String() string {
	return fmt.Sprintf("%d@%s%q", me.Sid,
		strings.ReplaceAll(me.When.Format(time.DateTime), " ", "T"),
		me.Comment)
}

func (me SaveInfoVal) marshal() ([]byte, error) {
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

func unmarshalSaveInfoVal(raw []byte) (SaveInfoVal, error) {
	var saveInfoVal SaveInfoVal
	if len(raw) == 0 {
		return saveInfoVal, errors.New("can't unmarshal empty saveval")
	}
	index := int(raw[0]) + 1
	var when time.Time
	if err := when.UnmarshalBinary(raw[1:index]); err != nil {
		return saveInfoVal, err
	}
	var comment string
	if len(raw) > index {
		comment = string(raw[index:])
	}
	saveInfoVal.When = when
	saveInfoVal.Comment = comment
	return saveInfoVal, nil
}
