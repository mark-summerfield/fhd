// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	rawFlag flag = iota
	flateFlag
	lzwFlag
)

type flag byte

func (me flag) String() string {
	switch me {
	case flateFlag:
		return "F"
	case lzwFlag:
		return "L"
	}
	return "R"
}

func flagForSizes(rawSize, flateSize, lzwSize int) flag {
	maxSize := int(float64(rawSize) * 0.95)
	if (flateSize > maxSize && lzwSize > maxSize) || (flateSize == 0 &&
		lzwSize == 0) {
		return rawFlag
	}
	if flateSize > 0 && flateSize < maxSize && (lzwSize == 0 ||
		(lzwSize > 0 && flateSize < lzwSize)) {
		return flateFlag
	}
	if lzwSize > 0 && lzwSize < maxSize {
		return lzwFlag
	}
	return rawFlag
}
