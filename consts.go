// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

type StateData struct {
	Sid      int // save ID
	Filename string
	State    StateKind
}

type StateKind uint8

const (
	Ignored StateKind = iota
	Unmonitored
	Monitored
)

const ModeOwnerRW = 0o600
