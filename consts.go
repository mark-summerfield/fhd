// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import _ "embed"

var (
	//go:embed Version.dat
	Version string

	FileFormat byte = 1

	ConfigBucket = []byte("config")
	StateBucket  = []byte("states")
	SavesBucket  = []byte("saves")

	ConfigFormat = []byte("format")
)

const ModeOwnerRW = 0o600
