package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExampleReset(t *testing.T) {
	// Create an instance with non-zero values
	counter := int64(42)
	score := float64(99.9)

	ex := &Example{
		ID:       123,
		Name:     "test",
		Active:   true,
		Tags:     []string{"tag1", "tag2", "tag3"},
		Metadata: map[string]string{"key1": "value1", "key2": "value2"},
		Counter:  &counter,
		Score:    &score,
		Nested: NestedStruct{
			Value: 456,
			Label: "nested",
		},
		PtrNest: &NestedStruct{
			Value: 789,
			Label: "pointer nested",
		},
	}

	// Save slice capacity before reset
	sliceCap := cap(ex.Tags)

	// Call Reset
	ex.Reset()

	// Verify all fields are reset to zero values
	assert.Equal(t, 0, ex.ID, "ID should be reset to 0")
	assert.Equal(t, "", ex.Name, "Name should be reset to empty string")
	assert.Equal(t, false, ex.Active, "Active should be reset to false")

	// Verify slice is truncated but capacity is preserved
	assert.Equal(t, 0, len(ex.Tags), "Tags length should be 0")
	assert.Equal(t, sliceCap, cap(ex.Tags), "Tags capacity should be preserved")

	// Verify map is cleared
	assert.Equal(t, 0, len(ex.Metadata), "Metadata should be empty")

	// Verify pointers to primitives are reset
	assert.NotNil(t, ex.Counter, "Counter pointer should not be nil")
	assert.Equal(t, int64(0), *ex.Counter, "Counter value should be 0")

	assert.NotNil(t, ex.Score, "Score pointer should not be nil")
	assert.Equal(t, float64(0), *ex.Score, "Score value should be 0")

	// Verify nested struct is reset
	assert.Equal(t, 0, ex.Nested.Value, "Nested.Value should be 0")
	assert.Equal(t, "", ex.Nested.Label, "Nested.Label should be empty")

	// Verify pointer to struct is reset
	assert.NotNil(t, ex.PtrNest, "PtrNest pointer should not be nil")
	assert.Equal(t, 0, ex.PtrNest.Value, "PtrNest.Value should be 0")
	assert.Equal(t, "", ex.PtrNest.Label, "PtrNest.Label should be empty")
}

func TestExampleResetWithNilPointers(t *testing.T) {
	// Create an instance with nil pointers
	ex := &Example{
		ID:       123,
		Name:     "test",
		Active:   true,
		Tags:     []string{"tag1"},
		Metadata: map[string]string{"key": "value"},
		Counter:  nil,
		Score:    nil,
		PtrNest:  nil,
	}

	// Call Reset
	ex.Reset()

	// Verify nil pointers remain nil (no panic)
	assert.Nil(t, ex.Counter, "Counter should remain nil")
	assert.Nil(t, ex.Score, "Score should remain nil")
	assert.Nil(t, ex.PtrNest, "PtrNest should remain nil")

	// Verify other fields are still reset
	assert.Equal(t, 0, ex.ID)
	assert.Equal(t, "", ex.Name)
	assert.Equal(t, 0, len(ex.Tags))
}

func TestComplexExampleResetWithNestedReset(t *testing.T) {
	// Create an instance with non-zero values including nested structs with Reset methods
	ptrInt := 999

	complex := &ComplexExample{
		ID:     42,
		Data:   []byte{1, 2, 3, 4},
		Items:  []string{"a", "b", "c"},
		Config: map[string]int{"x": 10, "y": 20},
		PtrInt: &ptrInt,
		SubStruct: SubWithReset{
			Name:  "sub1",
			Count: 100,
			Flags: []bool{true, false, true},
		},
		PtrSub: &SubWithReset{
			Name:  "sub2",
			Count: 200,
			Flags: []bool{false, true},
		},
	}

	// Save capacities before reset
	dataCap := cap(complex.Data)
	itemsCap := cap(complex.Items)
	subFlagsCap := cap(complex.SubStruct.Flags)
	ptrSubFlagsCap := cap(complex.PtrSub.Flags)

	// Call Reset
	complex.Reset()

	// Verify primitive fields
	assert.Equal(t, 0, complex.ID)
	assert.NotNil(t, complex.PtrInt)
	assert.Equal(t, 0, *complex.PtrInt)

	// Verify slices are truncated with preserved capacity
	assert.Equal(t, 0, len(complex.Data))
	assert.Equal(t, dataCap, cap(complex.Data))
	assert.Equal(t, 0, len(complex.Items))
	assert.Equal(t, itemsCap, cap(complex.Items))

	// Verify map is cleared
	assert.Equal(t, 0, len(complex.Config))

	// Verify SubStruct.Reset() was called (should have Reset method)
	assert.Equal(t, "", complex.SubStruct.Name)
	assert.Equal(t, 0, complex.SubStruct.Count)
	assert.Equal(t, 0, len(complex.SubStruct.Flags))
	assert.Equal(t, subFlagsCap, cap(complex.SubStruct.Flags))

	// Verify PtrSub.Reset() was called
	assert.NotNil(t, complex.PtrSub)
	assert.Equal(t, "", complex.PtrSub.Name)
	assert.Equal(t, 0, complex.PtrSub.Count)
	assert.Equal(t, 0, len(complex.PtrSub.Flags))
	assert.Equal(t, ptrSubFlagsCap, cap(complex.PtrSub.Flags))
}

func TestSubWithResetReset(t *testing.T) {
	// Test SubWithReset independently
	sub := &SubWithReset{
		Name:  "test",
		Count: 42,
		Flags: []bool{true, false, true, false},
	}

	flagsCap := cap(sub.Flags)

	sub.Reset()

	assert.Equal(t, "", sub.Name)
	assert.Equal(t, 0, sub.Count)
	assert.Equal(t, 0, len(sub.Flags))
	assert.Equal(t, flagsCap, cap(sub.Flags))
}
