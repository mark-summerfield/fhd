// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "time"

// Marshal for Time is: time.Time.MarshalBinary()
func UnmarshalTime(raw []byte) (time.Time, error) {
	var t time.Time
	err := t.UnmarshalBinary(raw)
	return t, err
}
