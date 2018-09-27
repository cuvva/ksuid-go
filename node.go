package ksuid

import (
	"sync"
	"time"
)

var exportedNode *Node

func init() {
	var iid InstanceID
	var err error

	iid, err = NewDockerID()
	if err != nil {
		iid, err = NewHardwareID()
		if err != nil {
			iid, err = NewRandomID()
			if err != nil {
				panic(err)
			}
		}
	}

	exportedNode = NewNode(Production, iid)
}

// Production is the internal name for production ksuid, but is omitted
// during marshaling.
const Production = "prod"

// Node contains metadata used for ksuid generation for a specific machine.
type Node struct {
	Environment string

	InstanceID InstanceID

	ts  time.Time
	seq uint32
	mu  sync.Mutex
}

// NewNode returns a ID generator for the current machine.
func NewNode(environment string, instanceID InstanceID) *Node {
	return &Node{
		Environment: environment,

		InstanceID: instanceID,
	}
}

// Generate returns a new ID for the machine and resource configured.
func (n *Node) Generate(resource string) (id ID) {
	id.Environment = n.Environment
	id.Resource = resource
	id.InstanceID = n.InstanceID

	n.mu.Lock()

	ts := time.Now().UTC()
	if ts.Sub(n.ts) > 1*time.Second {
		n.ts = ts
		n.seq = 0
	} else {
		n.seq++
	}

	id.Timestamp = n.ts
	id.SequenceID = n.seq

	n.mu.Unlock()

	return
}

// SetEnvironment overrides the default environment name in the exported node.
// This will effect all invocations of the exported Generate function.
func SetEnvironment(environment string) {
	exportedNode.Environment = environment
}

// SetInstanceID overrides the default instance id in the exported node.
// This will effect all invocations of the Generate function.
func SetInstanceID(instanceID InstanceID) {
	exportedNode.InstanceID = instanceID
}

// Generate returns a new ID for the current machine and resource configured.
func Generate(resource string) ID {
	return exportedNode.Generate(resource)
}
