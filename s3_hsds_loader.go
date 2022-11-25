// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3HSDSDomainLoader is an implementation of the HSDSDomainLoader,
// HSDSDomainVersionsLoader, and the HSDSObjectLoader interfaces that uses an S3
// bucket as its underlying storage.
type s3HSDSDomainLoader struct {
	// Client is the AWS S3 client used to send requests to the AWS S3 API.
	Client *s3.Client
	// Bucket is the bucket from which domains and domain objects are retrieved.
	Bucket string
}

func (l *s3HSDSDomainLoader) jsonForKey(ctx context.Context, key string, o interface{}) error {
	obj, err := l.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(l.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer obj.Body.Close()

	dec := json.NewDecoder(obj.Body)
	// For testing purposes, we fail on unknown fields. This should be removed
	// once everything is tested sufficiently.
	dec.DisallowUnknownFields()
	return dec.Decode(o)
}

func (l *s3HSDSDomainLoader) LoadDomain(ctx context.Context, name string) (*hsdsDomain, error) {
	p := path.Join(name, ".domain.json")
	d := &hsdsDomain{}
	err := l.jsonForKey(ctx, p, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (l *s3HSDSDomainLoader) LoadDomainVersions(ctx context.Context, domain *hsdsDomain) (map[string][]*hsdsVersion, error) {
	prefix := domain.DatabasePrefix()
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(l.Bucket),
		Prefix: aws.String(prefix),
	}
	output, err := l.Client.ListObjectVersions(ctx, input)
	if err != nil {
		return nil, err
	}

	versions := map[string][]*hsdsVersion{}
	for _, version := range output.Versions {
		key := aws.ToString(version.Key)
		vv, ok := versions[key]
		if !ok {
			vv = make([]*hsdsVersion, 0, 1)
		}
		v := &hsdsVersion{
			ID:           aws.ToString(version.VersionId),
			LastModified: aws.ToTime(version.LastModified),
			Size:         version.Size,
		}
		vv = append(vv, v)
		versions[key] = vv
	}

	// In theory, AWS should return the object versions sorted by their age
	// already, but better be safe than sorry.
	for _, ovs := range versions {
		sort.Slice(ovs, func(i, j int) bool {
			return ovs[i].LastModified.After(ovs[j].LastModified)
		})
	}

	return versions, nil
}

// ObjectForName loads the data associated with the object identified by key.
func (l *s3HSDSDomainLoader) LoadObject(ctx context.Context, name, version string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(l.Bucket),
		Key:    aws.String(name),
	}
	if version != "" {
		input.VersionId = aws.String(version)
	}

	obj, err := l.Client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()
	return ioutil.ReadAll(obj.Body)
}
