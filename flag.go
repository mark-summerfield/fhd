// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	rawFlag   flag = 'R'
	flateFlag flag = 'F'
	lzwFlag   flag = 'L'
)

type flag byte

func (me flag) String() string {
	return string(me)
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
