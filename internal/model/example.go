package model

// generate:reset
type Example struct {
	ID       int
	Name     string
	Active   bool
	Tags     []string
	Metadata map[string]string
	Counter  *int64
	Score    *float64
	Nested   NestedStruct
	PtrNest  *NestedStruct
}

type NestedStruct struct {
	Value int
	Label string
}

// generate:reset
type ComplexExample struct {
	ID        int
	Data      []byte
	Items     []string
	Config    map[string]int
	PtrInt    *int
	SubStruct SubWithReset
	PtrSub    *SubWithReset
}

// generate:reset
type SubWithReset struct {
	Name  string
	Count int
	Flags []bool
}
