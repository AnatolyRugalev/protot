package protoxtest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AnatolyRugalev/protox/testdata"
)

type T struct {
	errors []string
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
}

func TestEqual(t *testing.T) {
	testingT := &T{}
	assert := NewAssertions(testingT)
	expected := &testdata.A{
		Field1: "expected",
		Field2: true,
		Field3: 10,
	}
	actual := &testdata.A{
		Field1: "actual",
		Field2: false,
		Field3: 11,
	}
	require.False(t, assert.Equal(expected, actual))
	require.Len(t, testingT.errors, 1)
	require.Contains(t, testingT.errors[0], `Not equal:
	            	@ ["field1"]
	            	- "expected"
	            	+ "actual"
	            	@ ["field2"]
	            	- true
	            	@ ["field3"]
	            	- "10"
	            	+ "11"
`)
}

func TestSubset(t *testing.T) {
	testingT := &T{}
	assert := NewAssertions(testingT)
	a := &testdata.A{
		Field1: "a",
		Field2: true,
		Field3: 10,
	}
	b := &testdata.A{
		Field1: "a",
	}
	require.True(t, assert.Subset(a, b), testingT.errors)
}
