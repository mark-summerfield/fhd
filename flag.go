// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	Raw Flag = iota
	Patch
	Gz
	InOld
)

type Flag byte

func flagForSizes(rawSize, gzSize, patchSize int) Flag {
	frawSize := float64(rawSize)
	if (patchSize == 0 || gzSize < patchSize) &&
		gzSize < int(frawSize*0.95) {
		return Gz
	} else if patchSize > 0 && patchSize < int(frawSize*0.9) {
		return Patch
	}
	return Raw
}
