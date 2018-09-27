package ksuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"time"

	"github.com/jamescun/basex"
	"gopkg.in/mgo.v2/bson"
)

// ID is an optionally prefixed, k-sortable globally unique ID.
type ID struct {
	Environment string
	Resource    string

	Timestamp  time.Time
	InstanceID InstanceID
	SequenceID uint32
}

const (
	decodedLen = 21
	encodedLen = 29
)

// MustParse unmarshals an ID from a string and panics on error.
func MustParse(src string) ID {
	id, err := Parse([]byte(src))

	if err != nil {
		panic(err)
	}

	return id
}

// Parse unmarshals an ID from a series of bytes.
func Parse(src []byte) (id ID, err error) {
	id.Environment, id.Resource, src = splitPrefixID(src)

	if id.Environment == "" {
		id.Environment = Production
	}

	if len(src) < encodedLen {
		err = &ParseError{"ksuid too short"}
		return
	} else if len(src) > encodedLen {
		err = &ParseError{"ksuid too long"}
		return
	}

	dst := make([]byte, decodedLen)
	err = fastDecodeBase62(dst, src)
	if err != nil {
		err = &ParseError{"invalid base62: " + err.Error()}
		return
	}

	id.Timestamp = time.Unix(int64(binary.BigEndian.Uint64(dst[:8])), 0).UTC()
	id.InstanceID, err = ParseInstanceID(dst[8:17])
	id.SequenceID = binary.BigEndian.Uint32(dst[17:])

	return
}

func splitPrefixID(s []byte) (environment, resource string, id []byte) {
	// NOTE(jc): this function is optimized to reduce conditional branching
	// on the hot path/most common use case.

	i := bytes.LastIndexByte(s, '_')
	if i < 0 {
		id = s
		return
	}

	j := bytes.IndexByte(s[:i], '_')
	if j > -1 {
		environment = string(s[:j])
		resource = string(s[j+1 : i])
		id = s[i+1:]
		return
	}

	resource = string(s[:i])
	id = s[i+1:]

	return
}

// IsZero returns true if id has not yet been initialized.
func (id ID) IsZero() bool {
	return id.Environment == "" && id.Resource == "" &&
		id.Timestamp.IsZero() && id.InstanceID == nil &&
		id.SequenceID == 0
}

// Equal returns true if the given ID matches id of the caller.
func (id ID) Equal(x ID) bool {
	if id.InstanceID == nil || x.InstanceID == nil {
		return false
	}

	return id.Environment == x.Environment && id.Resource == x.Resource &&
		id.InstanceID.Scheme() == x.InstanceID.Scheme() &&
		id.InstanceID.Bytes() == x.InstanceID.Bytes() &&
		id.Timestamp.Equal(x.Timestamp) && id.SequenceID == x.SequenceID
}

// Scan implements a custom database/sql.Scanner to support
// unmarshaling from standard database drivers.
func (id *ID) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		n, err := Parse([]byte(src))
		if err != nil {
			return err
		}

		*id = n
		return nil

	case []byte:
		n, err := Parse(src)
		if err != nil {
			return err
		}

		*id = n
		return nil

	default:
		return &ParseError{"unsupported scan, must be string or []byte"}
	}
}

// Value implements a custom database/sql/driver.Valuer to support
// marshaling to standard database drivers.
func (id ID) Value() (driver.Value, error) {
	return id.Bytes(), nil
}

func (id ID) prefixLen() (n int) {
	if id.Resource != "" {
		n += len(id.Resource) + 1

		if id.Environment != "" && id.Environment != Production {
			n += len(id.Environment) + 1
		}
	}

	return
}

// MarshalJSON implements a custom JSON string marshaler.
func (id ID) MarshalJSON() ([]byte, error) {
	b := id.Bytes()
	x := make([]byte, len(b)+2)
	x[0] = '"'
	copy(x[1:], b)
	x[len(x)-1] = '"'
	return x, nil
}

// UnmarshalJSON implements a custom JSON string unmarshaler.
func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) < encodedLen+2 {
		return &ParseError{"ksuid too short"}
	} else if b[0] != '"' || b[len(b)-1] != '"' {
		return &ParseError{"expected string"}
	}

	n, err := Parse(b[1 : len(b)-1])
	if err != nil {
		return err
	}

	*id = n
	return nil
}

// GetBSON provides the necessary support for mgo.Getter
func (id ID) GetBSON() (interface{}, error) {
	return id.String(), nil
}

// GetBSON provides the necessary support for mgo.Setter
func (id *ID) SetBSON(raw bson.Raw) error {
	if raw.Kind != 0x02 {
		return &ParseError{"expected string"}
	}

	var str string
	if err := raw.Unmarshal(&str); err != nil {
		return err
	}

	n, err := Parse([]byte(str))
	if err != nil {
		return err
	}

	*id = n
	return nil
}

// Bytes stringifies and returns ID as a byte slice.
func (id ID) Bytes() []byte {
	prefixLen := id.prefixLen()
	dst := make([]byte, prefixLen+encodedLen)

	if id.Resource != "" {
		offset := 0
		if id.Environment != "" && id.Environment != Production {
			copy(dst, id.Environment)
			dst[len(id.Environment)] = '_'
			offset = len(id.Environment) + 1
		}

		copy(dst[offset:], id.Resource)
		dst[offset+len(id.Resource)] = '_'
	}

	iid := id.InstanceID.Bytes()

	x := make([]byte, decodedLen)
	y := make([]byte, encodedLen)
	binary.BigEndian.PutUint64(x, uint64(id.Timestamp.Unix()))
	x[8] = id.InstanceID.Scheme()
	copy(x[9:], iid[:])
	binary.BigEndian.PutUint32(x[17:], id.SequenceID)

	basex.Base62.Encode(y, x)
	copy(dst[prefixLen+2:], y)

	dst[prefixLen] = '0'
	dst[prefixLen+1] = '0'

	return dst
}

// String stringifies and returns ID as a string.
func (id ID) String() string {
	return string(id.Bytes())
}

// ParseError is returned when unexpected input is encountered when
// parsing user input to an ID.
type ParseError struct {
	errorString string
}

func (pe ParseError) Error() string {
	return pe.errorString
}
