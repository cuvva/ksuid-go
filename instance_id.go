package ksuid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// InstanceID is an interface implemented to identify a instance
// for a node in a unique manor.
type InstanceID interface {
	// Scheme returns the single byte used to identify the InstanceID.
	Scheme() byte

	// Bytes returns the serialized form of the InstanceID.
	Bytes() [8]byte
}

// ParseInstanceID unmarshals a prefixed node ID into its dedicated type.
func ParseInstanceID(b []byte) (InstanceID, error) {
	if len(b) != 9 {
		return nil, fmt.Errorf("expected 9 bytes, got %d", len(b))
	}

	switch b[0] {
	case 'H':
		return ParseHardwareID(b[1:])

	case 'D':
		return ParseDockerID(b[1:])

	case 'R':
		return ParseRandomID(b[1:])

	default:
		return nil, fmt.Errorf("unknown node id '%c'", b[0])
	}
}

// HardwareID identifies a Node using its Mac Address and Process ID.
type HardwareID struct {
	MachineID net.HardwareAddr
	ProcessID uint16
}

// NewHardwareID returns a HardwareID for the current node.
func NewHardwareID() (*HardwareID, error) {
	hwAddr, err := getHardwareAddr()
	if err != nil {
		return nil, err
	}

	return &HardwareID{
		MachineID: hwAddr,
		ProcessID: uint16(os.Getpid()),
	}, nil
}

func getHardwareAddr() (net.HardwareAddr, error) {
	addrs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// only return physical interfaces (i.e. not loopback)
		if len(addr.HardwareAddr) >= 6 {
			return addr.HardwareAddr, nil
		}
	}

	return nil, fmt.Errorf("no hardware addr available")
}

// ParseHardwareID unmarshals a HardwareID from a sequence of bytes.
func ParseHardwareID(b []byte) (*HardwareID, error) {
	if len(b) != 8 {
		return nil, fmt.Errorf("expected 8 bytes, got %d", len(b))
	}

	return &HardwareID{
		MachineID: net.HardwareAddr(b[:6]),
		ProcessID: binary.BigEndian.Uint16(b[6:]),
	}, nil
}

func (hid *HardwareID) Scheme() byte {
	return 'H'
}

func (hid *HardwareID) Bytes() [8]byte {
	var b [8]byte
	copy(b[:], hid.MachineID)
	binary.BigEndian.PutUint16(b[6:], hid.ProcessID)

	return b
}

// DockerID identifies a Node by its Docker container ID.
type DockerID struct {
	ContainerID []byte
}

// NewDockerID returns a DockerID for the current Docker container.
func NewDockerID() (*DockerID, error) {
	cid, err := getDockerID()
	if err != nil {
		return nil, err
	}

	return &DockerID{
		ContainerID: cid,
	}, nil
}

func getDockerID() ([]byte, error) {
	src, err := ioutil.ReadFile("/proc/1/cpuset")
	src = bytes.TrimSpace(src)
	if os.IsNotExist(err) || len(src) < 64 || !bytes.HasPrefix(src, []byte("/docker")) {
		return nil, fmt.Errorf("not a docker container")
	} else if err != nil {
		return nil, err
	}

	dst := make([]byte, 32)
	_, err = hex.Decode(dst, src[len(src)-64:])
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// ParseDockerID unmarshals a DockerID from a sequence of bytes.
func ParseDockerID(b []byte) (*DockerID, error) {
	if len(b) != 8 {
		return nil, fmt.Errorf("expected 8 bytes, got %d", len(b))
	}

	return &DockerID{
		ContainerID: b,
	}, nil
}

func (did *DockerID) Scheme() byte {
	return 'D'
}

func (did *DockerID) Bytes() [8]byte {
	var b [8]byte
	copy(b[:], did.ContainerID)

	return b
}

// RandomID identifies a Node by a random sequence of bytes.
type RandomID struct {
	Random [8]byte
}

// NewRandomID returns a RandomID initialized by a PRNG.
func NewRandomID() (*RandomID, error) {
	tmp := make([]byte, 8)
	rand.Read(tmp)

	var b [8]byte
	copy(b[:], tmp)

	return &RandomID{
		Random: b,
	}, nil
}

// ParseRandomID unmarshals a RandomID from a sequence of bytes.
func ParseRandomID(b []byte) (*RandomID, error) {
	if len(b) != 8 {
		return nil, fmt.Errorf("expected 8 bytes, got %d", len(b))
	}

	var x [8]byte
	copy(x[:], b)

	return &RandomID{
		Random: x,
	}, nil
}

func (rid *RandomID) Scheme() byte {
	return 'R'
}

func (rid *RandomID) Bytes() [8]byte {
	return rid.Random
}
