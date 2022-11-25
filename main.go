// Copyright 2022 UL Method Park GmbH. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main implements hss3dump.
//
// Hss3dump is a command-line tool for replicating a HSDS domain to your local
// filesystem. It will replicate it in a way that it can be used as the root
// directory for a local HSDS instance.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func usage() {

	fmt.Fprintf(os.Stderr, `usage: %s [OPTIONS] BUCKET DOMAIN...

Hss3dump downloads one or more HSDS domains from an S3 bucket, storing them on
the local filesystem in such a way that the target directory can be used as the
root directory for a local HSDS deployment.

It can restore different states of the target domain based on the versions
available in the S3 bucket. If an RFC3339 timestamp is supplied with the -b
flag, hss3dump will download the most recent versions of a domain's files that
are older or equal to the supplied time.

Options:
`, os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

func newS3Client() *s3.Client {
	conf, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		die(err)
	}
	client := s3.NewFromConfig(conf)
	return client
}

func list(bucket string, domains []string) {
	client := newS3Client()
	loader := &s3HSDSDomainLoader{
		Client: client,
		Bucket: bucket,
	}
	for _, name := range domains {
		domain, err := loader.LoadDomain(context.Background(), name)
		if err != nil {
			die(err)
		}
		versions, err := loader.LoadDomainVersions(context.Background(), domain)
		if err != nil {
			die(err)
		}

		fmt.Printf("%s:\n", name)
		objects := map[string][]byte{}
		for key, objectVersions := range versions {
			fmt.Printf("    %s\n", key)
			for _, version := range objectVersions {
				fmt.Printf("        %s\t%d Bytes\t%s\t\n",
					version.ID, version.Size, version.LastModified.Local().Format(time.RFC3339))
			}

			data, err := loader.LoadObject(context.Background(), key, "")
			if err != nil {
				die(err)
			}
			objects[key] = data
		}
		fmt.Println()
	}
}

// versionBefore returns the ID of the first version that is older than
// notAfter. It assumes that availableVersions is sorted by the versions'
// last modification time in descending order.
//
// If no version satisfies this condition the oldest version is returned.
// If not after is the zero value, the latest version is returned.
func versionBefore(availableVersions []*hsdsVersion, notAfter time.Time) string {
	if len(availableVersions) == 0 {
		panic("versionBefore: no versions available")
	}
	if notAfter.IsZero() {
		return availableVersions[0].ID
	}

	for _, version := range availableVersions {
		lm := version.LastModified.Local()
		if lm.Equal(notAfter) || lm.Before(notAfter) {
			return version.ID
		}
	}

	return availableVersions[len(availableVersions)-1].ID
}

func replicate(bucket, root string, domains []string, notAfter time.Time) {
	client := newS3Client()
	loader := &s3HSDSDomainLoader{
		Client: client,
		Bucket: bucket,
	}
	storer := &filesystemHSDSStorer{
		Root: root,
	}
	for _, name := range domains {
		domain, err := loader.LoadDomain(context.Background(), name)
		if err != nil {
			die(err)
		}
		ovs, err := loader.LoadDomainVersions(context.Background(), domain)
		if err != nil {
			die(err)
		}
		objectVersions := map[string]string{}
		for name, vv := range ovs {
			objectVersions[name] = versionBefore(vv, notAfter)
		}

		objects := map[string][]byte{}
		for name, version := range objectVersions {
			data, err := loader.LoadObject(context.Background(), name, version)
			if err != nil {
				die(err)
			}
			objects[name] = data
		}

		err = storer.StoreDomain(context.Background(), name, domain)
		if err != nil {
			die(err)
		}
		for name, b := range objects {
			err = storer.StoreObject(context.Background(), name, b)
			if err != nil {
				die(err)
			}
		}
	}
}

func main() {
	flag.Usage = usage

	var root string
	flag.StringVar(&root, "r", ".",
		"Choose the root directory of the local HSDS filesystem.")
	var before string
	flag.StringVar(&before, "b", "",
		"Return the first version of the domain before the given RFC3339 timestamp.")
	var cmdList bool
	flag.BoolVar(&cmdList, "l", false,
		"Output a list with all available file versions of each domain's files.")
	var help bool
	flag.BoolVar(&help, "h", false,
		"Print this command information.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if flag.NArg() < 2 {
		flag.Usage()
		return
	}

	args := flag.Args()
	bucket := args[0]
	domains := args[1:]
	if cmdList {
		list(bucket, domains)
	} else {
		var t time.Time
		var err error
		if before != "" {
			t, err = time.ParseInLocation(time.RFC3339, before, time.Local)
			if err != nil {
				die(err)
			}
		}
		replicate(bucket, root, domains, t)
	}
}
