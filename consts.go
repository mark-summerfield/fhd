// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import _ "embed"

var (
	//go:embed Version.dat
	Version string

	StateBucket   = []byte("states")
	SavesBucket   = []byte("saves")
	RenamedBucket = []byte("renamed")
)

const ModeOwnerRW = 0o600
