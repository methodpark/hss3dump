# hss3dump - Dump HSDS Domains to local filesystem

> **Important**: hss3dump is still in an early stage of development, so results
> may vary. It has only been tested with small datasets.

`hss3dump` is a command-line utility allowing you to list and/or replicate one
or more HSDS domains from an S3 bucket to your local filesystem. It will
replicate the data in such a way that it can be used as the root directory for a
local HSDS instance and, thus, allows restoring of h5 files using h5pyd's
`hsget`.

Additionally, if your S3 bucket has **versioning enabled** and its life cycle is
set up in such a way that it **does not delete** versions, the tool also
supports fetching different S3 object versions in order to allow restoring HSDS
domains from older versions.

## Installation

```sh
go install github.com/methodpark/hss3dump
```

## Usage

Hss3dump offers the `-h` flag to get more information on its usage:

```
$ hss3dump -h
usage: hss3dump [OPTIONS] BUCKET DOMAIN...

Hss3dump downloads one or more HSDS domains from an S3 bucket, storing them on
the local filesystem in such a way that the target directory can be used as the
root directory for a local HSDS deployment.

It can restore different states of the target domain based on the versions
available in the S3 bucket. If an RFC3339 timestamp is supplied with the -b
flag, hss3dump will download the most recent versions of a domain's files that
are older or equal to the supplied time.

Options:
  -b string
        Return the first version of the domain before the given RFC3339 timestamp.
  -h    Print this command information.
  -l    Output a list with all available file versions of each domain's files.
  -r string
        Choose the root directory of the local HSDS filesystem. (default ".")
```

### Fetching Most Recent Data

In order to fetch the most recent version of an HSDS domain called
`home/user/domain.h5` from an S3 bucket called `hsds-bucket` and dump it to the
current directory, run the following command:

```sh
$ hss3dump hsds-bucket home/user/domain.h5
```

### Supplying a Different Target Directory

The directory to which files will be written can be changed by specifying the
directory root with `-r`. Assuming we would like to replicate the domain above
to the directory `/var/db/hsds_data` the command would have to look like this:


```sh
$ hss3dump -r /var/db/hsds_data hsds-bucket home/user/domain.h5
```

### Restoring Previous Domain Versions

If we want to restore a previous version of a domain, we have to take a look at
the available versions first. Hss3dump makes this easy with its `-l` flag.
Assuming we have accidentally deleted data from a domain, the output could look
something like this:

```sh
$ hss3dump -l hsds-bucket home/user/domain.h5
home/user/domain.h5:
    db/e32b20a5-6c27622f/d/693e-302825-f8c087/.dataset.json
        ock7uFraVWjrotdTtGwXFR1N0TasC+ln        489 Bytes      2022-10-05T16:06:57+0100
    db/e32b60a5-6c27622f/d/693e-302825-f8c087/0
        HikS0B1PNyvCKLO+BmagsRaAnF1sL9zL        0 Bytes        2022-10-10 09:06:59+0100
        U9LG1wDd4EdzQj0PtZqPvvTH9/BdzvVH        1296 Bytes     2022-10-05 16:06:59+0100
    db/e32b60a5-6c27622f/g/40c5-5e41ac-92006c/.group.json
        sQwXZJAcjr1M0do1BsaFmnN6FlDLRwzM        1056 Bytes     2022-10-10 09:07:00+0100
        zkRK4cagD9alQWUeN3BKTi9T+SqQdjcO        193 Bytes      2022-10-05 16:07:00+0100
```

The output shows that the most recent version
(`HikS0B1PNyvCKLO+BmagsRaAnF1sL9zL`) of the file
`db/e32b60a5-6c27622f/d/693e-302825-f8c087/0` is 0 bytes large, while its
previous version was 1296 bytes large. This means that its content has been
deleted on the 10th October. If we want to restore the old data, we can do so by
letting hss3dump know that it should download only versions older than October
10th.  This can be done by supplying a corresponding RFC3339 timestamp via the
command's `-b` flag:

```sh
$ hss3dump -b "2022-10-10T00:00:00+0100" hsds-bucket home/user/domain.h5
```

Hss3dump will then either download the most recent version that satisfies this
condition, or - if no version of an object satisfies the condition - the oldest
version present is chosen instead.
