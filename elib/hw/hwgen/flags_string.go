// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// generated by stringer -type=Flags; DO NOT EDIT

package main

import "fmt"

const _Flags_name = "IsBitfieldIsFunc"

var _Flags_index = [...]uint8{0, 10, 16}

func (i Flags) String() string {
	i -= 1
	if i < 0 || i >= Flags(len(_Flags_index)-1) {
		return fmt.Sprintf("Flags(%d)", i+1)
	}
	return _Flags_name[_Flags_index[i]:_Flags_index[i+1]]
}