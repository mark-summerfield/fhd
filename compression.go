// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	noCompression    compression = 'U'
	flateCompression compression = 'F'
	lzwCompression   compression = 'L'
)

type compression byte

func (me compression) String() string {
	return string(me)
}

func compressionForSizes(rawSize, flateSize, lzwSize int) compression {
	maxSize := int(float64(rawSize) * 0.95)
	if (flateSize > maxSize && lzwSize > maxSize) || (flateSize == 0 &&
		lzwSize == 0) {
		return noCompression
	}
	if flateSize > 0 && flateSize < maxSize && (lzwSize == 0 ||
		(lzwSize > 0 && flateSize < lzwSize)) {
		return flateCompression
	}
	if lzwSize > 0 && lzwSize < maxSize {
		return lzwCompression
	}
	return noCompression
}
