package ksuid

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSplitPrefixID(t *testing.T) {
	tests := []struct {
		Name   string
		Source []byte

		Resource    string
		Environment string
		ID          []byte
	}{
		{"Empty", []byte(""), "", "", []byte("")},
		{"Bare", []byte("000EoVtOLK4o4XykFcYe63Kw"), "", "", []byte("000EoVtOLK4o4XykFcYe63Kw")},
		{"Resource", []byte("user_000EoVtOLK4o4XykFcYe63Kw"), "user", "", []byte("000EoVtOLK4o4XykFcYe63Kw")},
		{"ResourceEnvironment", []byte("test_user_000EoVtOLK4o4XykFcYe63Kw"), "user", "test", []byte("000EoVtOLK4o4XykFcYe63Kw")},
		{"BlankResource", []byte("_000EoVtOLK4o4XykFcYe63Kw"), "", "", []byte("000EoVtOLK4o4XykFcYe63Kw")},
		{"BlankResourceEnvironment", []byte("__000EoVtOLK4o4XykFcYe63Kw"), "", "", []byte("000EoVtOLK4o4XykFcYe63Kw")},
		{"BlankIDResource", []byte("user_"), "user", "", []byte("")},
		{"BlankIDResourceEnvironment", []byte("test_user_"), "user", "test", []byte("")},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			environment, resource, id := splitPrefixID(test.Source)

			assert.Equal(t, test.Environment, environment, "environment mismatch")
			assert.Equal(t, test.Resource, resource, "resource mismatch")
			assert.Equal(t, test.ID, id)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		Name   string
		Source []byte

		ID    ID
		Error error
	}{
		{"Short", []byte(""), ID{}, &ParseError{"ksuid too short"}},
		{"Long", []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), ID{}, &ParseError{"ksuid too long"}},
		{"InvalidBase62", []byte("AAAAAAAAAAAAAAAAAAAAAAAAA//AA"), ID{}, &ParseError{"invalid base62: output buffer too short"}},
		{
			"Bare", []byte("000000BPG6Lks9tQoAiJYrBRSXPX6"),
			ID{
				Environment: Production,
				Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
				InstanceID: &HardwareID{
					MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
					ProcessID: 32985,
				},
				SequenceID: 0,
			}, nil,
		},
		{
			"Resource", []byte("user_000000BPG6Lks9tQoAiJYrBRSXPX6"),
			ID{
				Environment: Production,
				Resource:    "user",
				Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
				InstanceID: &HardwareID{
					MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
					ProcessID: 32985,
				},
				SequenceID: 0,
			}, nil,
		},
		{
			"ResourceEnvironment", []byte("test_user_000000BPG6Lks9tQoAiJYrBRSXPX6"),
			ID{
				Environment: "test",
				Resource:    "user",
				Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
				InstanceID: &HardwareID{
					MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
					ProcessID: 32985,
				},
				SequenceID: 0,
			}, nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			id, err := Parse(test.Source)
			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.ID, id)
				}
			} else {
				assert.Equal(t, test.Error, err)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse([]byte("user_000000BPG6Lks9tQoAiJYrBRSXPX6"))
	}
}

func TestID(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		tests := []struct {
			Name string
			Src  interface{}

			ID    ID
			Error error
		}{
			{
				"Bytes", []byte("000000BPG6Lks9tQoAiJYrBRSXPX6"),
				ID{
					Environment: Production,
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, nil,
			},
			{
				"String", "000000BPG6Lks9tQoAiJYrBRSXPX6",
				ID{
					Environment: Production,
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, nil,
			},
			{
				"Unknown", 1234, ID{}, &ParseError{"unsupported scan, must be string or []byte"},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				id := ID{}
				err := id.Scan(test.Src)
				if test.Error == nil {
					if assert.NoError(t, err) {
						assert.Equal(t, test.ID, id)
					}
				} else {
					assert.Equal(t, test.Error, err)
				}
			})
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			Name   string
			Source []byte

			ID    ID
			Error error
		}{
			{"Short", []byte(`""`), ID{}, &ParseError{"ksuid too short"}},
			{"Long", []byte(`"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"`), ID{}, &ParseError{"ksuid too long"}},
			{"NotString", []byte(`1111111111111111111111111111111`), ID{}, &ParseError{"expected string"}},
			{"InvalidBase62", []byte(`"AAAAAAAAAAAAAAAAAAAAAAAAA//AA"`), ID{}, &ParseError{"invalid base62: output buffer too short"}},
			{
				"Bare", []byte(`"000000BPG6Lks9tQoAiJYrBRSXPX6"`),
				ID{
					Environment: Production,
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, nil,
			},
			{
				"Resource", []byte(`"user_000000BPG6Lks9tQoAiJYrBRSXPX6"`),
				ID{
					Resource:    "user",
					Environment: Production,
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, nil,
			},
			{
				"ResourceEnvironment", []byte(`"test_user_000000BPG6Lks9tQoAiJYrBRSXPX6"`),
				ID{
					Resource:    "user",
					Environment: "test",
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, nil,
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				id := ID{}
				err := id.UnmarshalJSON(test.Source)
				if test.Error == nil {
					if assert.NoError(t, err) {
						assert.Equal(t, test.ID, id)
					}
				} else {
					assert.Equal(t, test.Error, err)
				}
			})
		}
	})

	t.Run("Bytes", func(t *testing.T) {
		tests := []struct {
			Name string
			ID   ID

			Bytes []byte
			JSON  []byte
		}{
			{
				"Bare", ID{
					Timestamp: time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, []byte("000000BPG6Lks9tQoAiJYrBRSXPX6"), []byte(`"000000BPG6Lks9tQoAiJYrBRSXPX6"`),
			},
			{
				"BareEnvironment", ID{
					Environment: "test",
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, []byte("000000BPG6Lks9tQoAiJYrBRSXPX6"), []byte(`"000000BPG6Lks9tQoAiJYrBRSXPX6"`),
			},
			{
				"Resource", ID{
					Resource:  "user",
					Timestamp: time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, []byte("user_000000BPG6Lks9tQoAiJYrBRSXPX6"), []byte(`"user_000000BPG6Lks9tQoAiJYrBRSXPX6"`),
			},
			{
				"ResourceProduction", ID{
					Environment: Production,
					Resource:    "user",
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, []byte("user_000000BPG6Lks9tQoAiJYrBRSXPX6"), []byte(`"user_000000BPG6Lks9tQoAiJYrBRSXPX6"`),
			},
			{
				"ResourceEnvironment", ID{
					Resource:    "user",
					Environment: "test",
					Timestamp:   time.Date(2018, 4, 5, 16, 53, 42, 0, time.UTC),
					InstanceID: &HardwareID{
						MachineID: net.HardwareAddr{0x8C, 0x85, 0x90, 0x5F, 0x44, 0xCA},
						ProcessID: 32985,
					},
					SequenceID: 0,
				}, []byte("test_user_000000BPG6Lks9tQoAiJYrBRSXPX6"), []byte(`"test_user_000000BPG6Lks9tQoAiJYrBRSXPX6"`),
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				assert.Equal(t, test.Bytes, test.ID.Bytes(), "bytes mismatch")
				assert.Equal(t, string(test.Bytes), test.ID.String(), "string mismatch")

				value, err := test.ID.Value()
				if assert.NoError(t, err) {
					assert.Equal(t, test.Bytes, value)
				}

				json, err := test.ID.MarshalJSON()
				if assert.NoError(t, err) {
					assert.Equal(t, test.JSON, json)
				}
			})
		}
	})
}
