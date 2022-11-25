// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

var (
	validGroupID                      = hsdsID{'g', 0xd1, 0x2a, 0x20, 0xa5, 0x6c, 0x27, 0x62, 0x2f, 0x59, 0xa2, 0xa8, 0x2d, 0xe4, 0xaf, 0xea, 0xa7}
	validGroupIDString                = "g-d12a20a5-6c27622f-59a2-a82de4-afeaa7"
	invalidEntityType  hsdsEntityType = 'x'
	invalidID                         = hsdsID{byte(invalidEntityType)}
	invalidIDString                   = "%c-d12a20a5-6c27622f-59a2-a82de4-afeaa7"
)

type marshalTestcase struct {
	name    string
	id      hsdsID
	want    []byte
	wantErr error
}

func TestID_MarshalText(t *testing.T) {
	testCases := []marshalTestcase{
		{
			name:    "valid-group-id",
			id:      validGroupID,
			want:    []byte(validGroupIDString),
			wantErr: nil,
		},
		{
			name:    "invalid-hdf5-type",
			id:      invalidID,
			want:    nil,
			wantErr: &unknownEntityTypeError{Type: invalidEntityType},
		},
	}

	for _, tc := range testCases {
		got, err := tc.id.MarshalText()
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("%s: id.MarshalText() err = %v (want %v)", tc.name, err, tc.wantErr)
			return
		}
		if tc.wantErr != nil {
			return
		}

		if !bytes.Equal(got, tc.want) {
			t.Errorf("%s: id.MarshalText() = %q (want %q)", tc.name, got, tc.want)
		}
	}
}

type unmarshalTestcase struct {
	name    string
	id      []byte
	want    hsdsID
	wantErr error
}

func TestID_UnmarshalText(t *testing.T) {
	testCases := []unmarshalTestcase{
		{
			name:    "valid-id",
			id:      []byte(validGroupIDString),
			want:    validGroupID,
			wantErr: nil,
		},
		{
			name:    "invalid-hdf5-type",
			id:      []byte(fmt.Sprintf("%c-d12a20a5-6c27622f-59a2-a82de4-afeaa7", invalidEntityType)),
			wantErr: &unknownEntityTypeError{Type: invalidEntityType},
		},
	}

	for _, tc := range testCases {
		var got hsdsID
		err := got.UnmarshalText(tc.id)
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("%s: id.UnmarshalText() err = %v (want %v)", tc.name, err, tc.wantErr)
			return
		}
		if tc.wantErr != nil {
			return
		}

		if !bytes.Equal(got[:], tc.want[:]) {
			t.Errorf("%s: id.UnmarshalText() = %q (want %q)", tc.name, got, tc.want)
		}
	}
}
