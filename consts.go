// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import _ "embed"

var (
	//go:embed Version.dat
	Version string

	fileFormat byte = 1

	configBucket = []byte("config")
	stateBucket  = []byte("states")
	savesBucket  = []byte("saves")

	configFormat = []byte("format")
)

const modeUserRW = 0o600
