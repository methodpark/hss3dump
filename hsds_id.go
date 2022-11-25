// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"errors"
)

// errInvalidHSDSID indicates that the UUID portion of a parsed HSDSID is invalid.
var errInvalidHSDSID = errors.New("hsds: invalid HSDS UUID format")

// A HSDS id consists of a one byte HDF5 type, plus a 128 bit UUID.
type hsdsID [17]byte

// nilID is the zero value of an HSDSID.
var nilID = hsdsID{}

// ParseID trys to parse id and returns it if successful.
//
// On success an HSDSID corresponding to the parsed ID is returned. Otherwise,
// the NilID and an error indicating why the parsing operation has failed is
// returned.
func ParseID(id string) (hsdsID, error) {
	newID := hsdsID{}
	err := newID.UnmarshalText([]byte(id))
	if err != nil {
		return nilID, err
	}
	return newID, nil
}

// MustParseID is like ParseID, but it panics if an error occurs.
func MustParseID(id string) hsdsID {
	i, err := ParseID(id)
	if err != nil {
		panic(err)
	}
	return i
}

// Type returns the id's entity type.
func (id hsdsID) Type() hsdsEntityType {
	return hsdsEntityType(id[0])
}

const hsdsIDLen = 38

var (
	// Text-encoded HSDSIDs have the following format:
	// x-xxxxxxxx-xxxxxxxx-xxxx-xxxxxx-xxxxxx
	hsdsIDDashIndeces = []int{1, 10, 19, 24, 31}
	hsdsByteIndices   = []int{2, 4, 6, 8, 11, 13, 15, 17, 20, 22, 25, 27, 29, 32, 34, 36}
)

func (id hsdsID) MarshalText() ([]byte, error) {
	t := hsdsEntityType(id[0])
	if !t.Valid() {
		return nil, &unknownEntityTypeError{Type: t}
	}

	return []byte(id.String()), nil
}

func (id *hsdsID) UnmarshalText(b []byte) error {
	if len(b) != hsdsIDLen {
		return errInvalidHSDSID
	}
	t := hsdsEntityType(b[0])
	if !t.Valid() {
		return &unknownEntityTypeError{Type: t}
	}

	id[0] = b[0]
	dest := id[1:]
	for i, j := range hsdsByteIndices {
		_, err := hex.Decode(dest[i:i+1], b[j:j+2])
		if err != nil {
			return errInvalidHSDSID
		}
	}

	return nil
}

func (id hsdsID) String() string {
	b := make([]byte, hsdsIDLen)
	b[0] = id[0]
	src := id[1:]
	for i, j := range hsdsByteIndices {
		hex.Encode(b[j:j+2], src[i:i+1])
	}
	for _, i := range hsdsIDDashIndeces {
		b[i] = '-'
	}
	return string(b)
}

// hsdsPrefix is the type representing the id prefix for an ID. It consists of
// the first eight bytes of the ID's UUID.
type hsdsPrefix [8]byte

// Prefix returns the ID's HSDS prefix. The prefix is used to form paths to
// groups, committed types, datasets and chunks belonging to the same domain.
func (id hsdsID) Prefix() hsdsPrefix {
	p := hsdsPrefix{}
	copy(p[:], id[1:9])
	return p
}

const prefixLen = 17

func (p hsdsPrefix) String() string {
	// The prefix is of the form xxxxxxxx-xxxxxxxx
	b := make([]byte, prefixLen)
	hex.Encode(b[:8], p[:4])
	b[8] = '-'
	hex.Encode(b[9:], p[4:])

	return string(b)
}

// hsdsSuffix is the type representing the id suffix for an ID. It consists of the
// last eight bytes of the ID's UUID.
type hsdsSuffix [8]byte

func (id hsdsID) Suffix() hsdsSuffix {
	s := hsdsSuffix{}
	copy(s[:], id[9:])
	return s
}

const suffixLen = 18

func (s hsdsSuffix) String() string {
	// The suffix is of the form xxxx-xxxxxx-xxxxxx
	b := make([]byte, suffixLen)
	hex.Encode(b[:4], s[:2])
	b[4] = '-'
	hex.Encode(b[5:11], s[2:5])
	b[11] = '-'
	hex.Encode(b[12:], s[5:])

	return string(b)
}

// hsdsUUID is the type representing an IDs hsdsUUID portion. It consists of all
// bytes except the first.
type hsdsUUID [16]byte

// UUID returns an HSDSID's UUID portion
func (id hsdsID) UUID() hsdsUUID {
	uuid := hsdsUUID{}
	copy(uuid[:], id[1:])
	return uuid
}

const uuidLen = 36

var (
	// uuidByteIndices is an array mapping an UUID's bytes to byte positions in
	// its text encoded form.
	//
	// Text encoded UUIDs have the following form: xxxxxxxx-xxxxxxxx-xxxx-xxxxxx-xxxxxx
	uuidByteIndices = []int{0, 2, 4, 6, 9, 11, 13, 15, 18, 20, 23, 25, 27, 30, 32, 34}
	uuidDashIndices = []int{8, 17, 22, 29}
)

func (uuid hsdsUUID) MarshalText() ([]byte, error) {
	return []byte(uuid.String()), nil
}

var ErrInvalidUUID = errors.New("hsds: HSDSD id has nil UUID")

func (uuid *hsdsUUID) UnmarshalText(b []byte) error {
	if len(b) != uuidLen {
		return ErrInvalidUUID
	}

	for _, i := range uuidDashIndices {
		if b[i] != '-' {
			return ErrInvalidUUID
		}
	}
	for i, j := range uuidByteIndices {
		_, err := hex.Decode(uuid[i:i+1], b[j:j+2])
		if err != nil {
			return ErrInvalidUUID
		}
	}

	return nil
}

func (uuid hsdsUUID) String() string {
	b := make([]byte, uuidLen)
	for i, j := range uuidByteIndices {
		hex.Encode(b[j:j+2], uuid[i:i+1])
	}
	for _, i := range uuidDashIndices {
		b[i] = '-'
	}
	return string(b)
}
