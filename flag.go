// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	Raw Flag = iota
	Flate
	Lzw
)

type Flag byte

func (me Flag) String() string {
	switch me {
	case Flate:
		return "F"
	case Lzw:
		return "L"
	}
	return "R"
}

func flagForSizes(rawSize, flateSize, lzwSize int) Flag {
	maxSize := int(float64(rawSize) * 0.95)
	if (flateSize > maxSize && lzwSize > maxSize) || (flateSize == 0 &&
		lzwSize == 0) {
		return Raw
	}
	if flateSize > 0 && flateSize < maxSize && (lzwSize == 0 ||
		(lzwSize > 0 && flateSize < lzwSize)) {
		return Flate
	}
	if lzwSize > 0 && lzwSize < maxSize {
		return Lzw
	}
	return Raw
}
