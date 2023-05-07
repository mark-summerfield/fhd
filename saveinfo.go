// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "time"

type SaveInfo struct {
	Sid     SID
	When    time.Time
	Comment string
}

func newSaveInfo(sid SID, when time.Time, comment string) SaveInfo {
	return SaveInfo{Sid: sid, When: when, Comment: comment}
}

func newInvalidSaveInfo() SaveInfo {
	return SaveInfo{Sid: InvalidSID}
}

func (me *SaveInfo) IsValid() bool {
	return me.Sid.IsValid()
}
