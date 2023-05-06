// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "time"

type SID uint64

type SidInfo struct {
	sid     SID
	when    time.Time
	comment string
}

func newSidInfo(sid SID, when time.Time, comment string) SidInfo {
	return SidInfo{sid: sid, when: when, comment: comment}
}

func newInvalidSidInfo() SidInfo {
	return SidInfo{sid: 0}
}

func (me *SidInfo) IsValid() bool {
	return me.sid > 0
}

func (me *SidInfo) RawSid() []byte {
	return MarshalSid(me.sid)
}

func (me *SidInfo) Sid() SID {
	return me.sid
}

func (me *SidInfo) When() time.Time {
	return me.when
}

func (me *SidInfo) Comment() string {
	return me.comment
}
