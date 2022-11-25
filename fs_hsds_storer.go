// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type pathError struct {
	path string
}

func (err *pathError) Error() string {
	return fmt.Sprintf("filesystem: '%s' is not a valid filename", err.path)
}

// filesystemHSDSStorer is an implementation of the DomainStorer and
// ObjectStorer interfaces that uses the local filesystem as its underlying
// storage.
type filesystemHSDSStorer struct {
	// Root is the storer's root directory. All domains and domain objects
	// stored by the storer will reside in this directory.
	Root string
}

func sanitizePath(root, name string) (string, error) {
	name = filepath.Join("/", filepath.FromSlash(name))
	if name == "/" {
		return "", &pathError{path: name}
	}
	name = filepath.Join(root, name)
	return name, nil
}

func openForWriting(root, name string) (io.WriteCloser, error) {
	name, err := sanitizePath(root, name)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return f, err
}

func createParentDomains(root, name string, domain *hsdsDomain) error {
	name = filepath.Clean(name)
	if name == "." {
		return nil
	}

	dirName, err := sanitizePath(root, name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dirName, 0744)
	if err != nil {
		return err
	}

	parentDir, _ := filepath.Split(name)
	parentDirs := filepath.SplitList(parentDir)
	// Directory domains do not have a root group.
	parent := *domain
	parent.Root = nil
	dn := root
	for _, subDir := range parentDirs {
		dn = filepath.Join(dn, subDir, ".domain.json")
		f, err := os.OpenFile(dn, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
		// We only create domain files for parent directories that do not already exist.
		if errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return err
		}

		enc := json.NewEncoder(f)
		err = enc.Encode(parent)
		if err != nil {
			f.Close()
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *filesystemHSDSStorer) StoreDomain(ctx context.Context, name string, domain *hsdsDomain) error {
	err := createParentDomains(s.Root, name, domain)
	if err != nil {
		return err
	}

	name = filepath.Join(name, ".domain.json")
	f, err := openForWriting(s.Root, name)
	enc := json.NewEncoder(f)
	err = enc.Encode(domain)
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (s *filesystemHSDSStorer) StoreObject(ctx context.Context, name string, data []byte) error {
	dir, err := sanitizePath(s.Root, name)
	if err != nil {
		return err
	}
	dir, _ = filepath.Split(dir)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := openForWriting(s.Root, name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
