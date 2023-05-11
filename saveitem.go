// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"fmt"
	"strings"
	"time"
)

type SaveItem struct {
	Sid     SID
	When    time.Time
	Comment string
}

func newSaveItem(sid SID, when time.Time, comment string) SaveItem {
	return SaveItem{Sid: sid, When: when, Comment: comment}
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
