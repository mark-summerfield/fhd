// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import _ "embed"

var (
	//go:embed Version.dat
	Version string

	fileFormat byte = 1

	configBucket  = []byte("config")
	statesBucket  = []byte("states")
	savesBucket   = []byte("saves")
	renamedBucket = []byte("renamed")

	configFormat = []byte("format")
	configIgnore = []byte("ignore")

	saveWhen           = []byte("*when")
	saveComment        = []byte("*comment")
	savePredefinedKeys = 2 // how many of the above

	emptyValue = []byte{}

	// Should also ignore hidden (.) files and subdirs by default.
	defaultIgnores = []string{"*#[0-9].*", "*.a", "*.bak", "*.class",
		"*.dll", "*.exe", "*.fhd", "*.jar", "*.ld", "*.ldx", "*.li",
		"*.lix", "*.o", "*.obj", "*.py[co]", "*.rs.bk", "*.so", "*.sw[nop]",
		"*.swp", "*.tmp", "*~", "gpl-[0-9].[0-9].txt", "louti[0-9]*",
		"moc_*.cpp", "qrc_*.cpp", "ui_*.h"}

	uint16size = 2
)
