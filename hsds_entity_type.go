// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

// hsdsEntityType is a representation of HSDS object types.
type hsdsEntityType byte

const (
	entityTypeGroup         hsdsEntityType = 'g'
	entityTypeDataset       hsdsEntityType = 'd'
	entityTypeCommittedType hsdsEntityType = 't'
)

// Valid returns, whether t is a valid entity type.
func (t hsdsEntityType) Valid() bool {
	return (t == entityTypeGroup || t == entityTypeDataset || t == entityTypeCommittedType)
}

// unknownEntityTypeError is an error indicating that the an unknown entity type
// has been encountered.
type unknownEntityTypeError struct {
	Type hsdsEntityType
}

func (err *unknownEntityTypeError) Error() string {
	return fmt.Sprintf("hsds: unknown HDF5 type '%c'", err.Type)
}

func (err *unknownEntityTypeError) Is(other error) bool {
	x, ok := other.(*unknownEntityTypeError)
	if !ok {
		return false
	}
	return x.Type == err.Type
}
