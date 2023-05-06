// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "time"

const InvalidSID = 0

type SID uint32 // allows for 4 billion saves

func (me SID) IsValid() bool { return me != InvalidSID }

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

func (me *SaveInfo) RawSid() []byte {
	return MarshalSid(me.Sid)
}
