// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"path"
	"time"
)

// hsdsACL is the Access Control List for an HSDS domain.
type hsdsACL map[string]*hsdsPermissions

// hsdsPermissions are the permissions for a single user.
type hsdsPermissions struct {
	Create    bool `json:"create"`
	Read      bool `json:"read"`
	Update    bool `json:"update"`
	Delete    bool `json:"delete"`
	ReadACL   bool `json:"readACL"`
	UpdateACL bool `json:"updateACL"`
}

// hsdsDomain is roughly the equivalent of an HDF5 file in an S3 bucket.
type hsdsDomain struct {
	ACLs         hsdsACL `json:"acls"`
	Root         *hsdsID `json:"root,omitempty"`
	Owner        string  `json:"owner"`
	Created      float64 `json:"created,omitempty"`
	LastModified float64 `json:"lastModified,omitempty"`
}

// Prefix returns d's root group's ID prefix.
func (d *hsdsDomain) Prefix() hsdsPrefix {
	return d.Root.Prefix()
}

// Suffix returns d's root group's ID suffix.
func (d *hsdsDomain) Suffix() hsdsSuffix {
	return d.Root.Suffix()
}

// DatabasePrefix returns the path prefix for all objects in a HSDS-based
// database that belong to d.
func (d *hsdsDomain) DatabasePrefix() string {
	return path.Join("db", d.Prefix().String())
}

// hsdsDomainLoader is the interface implementing the LoadDomain method.
//
// LoadDomain loads the domain identified by name in the loaders's persistent
// storage.
type hsdsDomainLoader interface {
	LoadDomain(ctx context.Context, name string) (*hsdsDomain, error)
}

// hsdsDomainStorer is the interface implementing the StoreDomain method.
//
// StoreDomain stores domain under the given name in the storer's persistent
// storage.
type hsdsDomainStorer interface {
	StoreDomain(ctx context.Context, name string, domain *hsdsDomain) error
}

// HSDSLoadStorer is the combination of the HSDSDomainLoader and
// HSDSDomainStorer interfaces.
type hsdsDomainLoadStorer interface {
	hsdsDomainLoader
	hsdsDomainStorer
}

// hsdsVersion is the type representing a specific version of a domain object.
type hsdsVersion struct {
	ID           string
	LastModified time.Time
	Size         int64
}

// hsdsDomainVersionLoader wraps the LoadDomainVersions method.
//
// LoadDomainVersions loads all domain's object versions.
//
// On success a map is returned, where each key-value pair consists of the path
// identifying the domain object and the respective object's versions. Otherwise,
// nil and an error is returned.
type hsdsDomainVersionLoader interface {
	LoadDomainVersions(ctx context.Context, domain *hsdsDomain) map[string][]*hsdsVersion
}

// hsdsObjectLoader is the interface wrapping the LoadObject method.
//
// LoadObjects loads the givne version of the domain object from the loader's
// underlying persistent storage.
//
// On success, it returns the data associated with the domain objects.
// Otherwise a nil and an appropriate error is returned.
type hsdsObjectLoader interface {
	LoadObject(ctx context.Context, name, version string) ([]byte, error)
}

// hsdsObjectStorer is the interface wrapping the StoreObjects method.
//
// StoreObject stores data under the given path in the storer's underlying
// persistent storage.
//
// On success nil is returned. Otherwise, an error indicating the cause of
// failure is returned.
type hsdsObjectStorer interface {
	StoreObject(ctx context.Context, name string, data []byte) error
}
