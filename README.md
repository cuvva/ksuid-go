# ksuid

ksuid is a Go library that generated prefixed, k-sorted globally unique identifiers.

Each ksuid has a resource type and optionally an environment prefix (no environment prefix is for production use only). They are roughly sortable down to per-second resolution.

Properties of a ksuid:

  - resource type and environment prefixing
  - lexicographically, time sortable
  - no startup co-ordination
  - guaranteed unique relative to process/machine

## Usage

### API

ksuid is primarily a Go package to be consumed by Cuvva services, below are examples of its API usage.

To generate a ksuid with a custom resource type and for the production environment:

```go
id := ksuid.Generate("user")
/* => ID{
	Environment: "prod",
	Resource: "user",
	Timestamp: time.Time{"2018-09-27T15:44:10Z"},
	MachineID: net.HardwareAddr{"8c:85:90:b3:c9:0f"},
	ProcessID: 21086,
	SequenceID: 0,
} */
```

To parse a single given ksuid:

```go
id, err := ksuid.Parse([]byte("user_000000BWHKXYBQe5Dt06dsPlgJAh6"))
/*
=> ID{
	Environment: "prod",
	Resource: "user",
	Timestamp: time.Time{"2018-09-27T15:44:10Z"},
	MachineID: net.HardwareAddr{"8c:85:90:b3:c9:0f"},
	ProcessID: 21086,
	SequenceID: 0,
}, nil
*/
```

### Command Line Tool

```sh
go get -u github.com/cuvva/ksuid-go/cmd/ksuid
```

ksuid provides a helper utility to generate and parse ksuid on the command line, it contains two subcommands: `parse` and `generate`.

To generate two ksuid with a custom resource type and for the production environment:

```sh
$ ksuid generate --resource=user --count=2
user_000000BWHKXYBQe5Dt06dsPlgJAh6
user_000000BWHL2JrcCYZwWda13ERyoam
```

To parse a single given ksuid:

```sh
$ ksuid parse user_000000BWHKXYBQe5Dt06dsPlgJAh6
ID:          user_000000BWHKXYBQe5Dt06dsPlgJAh6
Resource:    user
Environment: prod
Timestamp:   2018-09-27T15:44:10Z
Machine ID:  8c:85:90:b3:c9:0f
Process ID:  21086
Sequence ID: 0
```

## How They Work

ksuid are minimum 22 bytes long when Base62 encoded, consisting of 16 bytes decoded:

  - a 32-bit unix timestamp with a custom epoch of 2014-01-01T00:00:00Z
  - 9 bytes of data that is unqiue per-machine
  - the 16-bit process id of the generating service
  - a 32-bit incrementing counter, reset every second

Optionally a ksuid has two, underscore delimited prefixes. The first prefix is optional, and is the environment in which the ksuid was generated (test, dev, git commit etc), omitting the environment identifies production only. The second prefix is the resource type (user, profile, vehicle etc) and is required.

For more information, [check out our blog post](https://www.cuvva.com/car-insurance/showing-off-our-fancy-new-ids/) on how we designed ksuid.
