package ksuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetIterator(t *testing.T) {
	tests := []struct {
		Name string
		Set  []ID
	}{
		{"Empty", []ID{}},
		{"Single", []ID{Generate("example")}},
		{"Multiple", []ID{Generate("example"), Generate("example")}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s := NewSet(test.Set...)
			assert.Equal(t, len(test.Set), s.Len())

			count := 0
			iter := s.Iter()

			for iter.Next() {
				assert.True(t, test.Set[count].Equal(iter.Value()), "id mismatch")

				count++
			}

			assert.Equal(t, len(test.Set), count, "item count mismatch")
		})
	}
}
