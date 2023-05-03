// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	Raw Flag = iota
	Flate
	Lzw
)

type Flag byte

func flagForSizes(rawSize, flateSize, lzwSize int) Flag {
	maxSize := int(float64(rawSize) * 0.95)
	if flateSize > maxSize && lzwSize > maxSize {
		return Raw
	}
	if flateSize < lzwSize {
		return Flate
	}
	return Lzw
}
