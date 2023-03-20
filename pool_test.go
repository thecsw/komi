package komi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoolCreationClosing(t *testing.T) {
	simplePool := NewPool(WorkSimple(squareSimple),
		WithDebug(), WithName("Simple Pool"), WithLaborers(10))
	assert.NotNil(t, simplePool, "simple pool")
	simplePool.Close()
	assert.Equal(t, true, simplePool.closed, "closed flag")
	assert.NotEmpty(t, simplePool.closureSignalForChildren, "closed signal channel")

	simplePoolWithErrors := NewPool(WorkSimpleWithErrors(squareSimpleWithErrors),
		WithDebug(), WithName("Simple Pool With Errors"))
	assert.NotNil(t, simplePoolWithErrors, "simple pool with errors")

	regularPool := NewPool(Work(squareRegular),
		WithDebug(), WithName("Regular Pool"))
	assert.NotNil(t, regularPool, "regular pool")

	regularPoolWithErrors := NewPool(WorkWithErrors(squarReguralWithErrors),
		WithDebug(), WithName("Regular Pool With Errors"))
	assert.NotNil(t, regularPoolWithErrors, "regular pool with errors")
}

func squareSimple(v int) {
	v *= v
}

func squareSimpleWithErrors(v int) error {
	if v <= 0 {
		return errors.New("only positives allowed")
	}
	return nil
}

func squareRegular(v int) int {
	return v * v
}

func squarReguralWithErrors(v int) (int, error) {
	if v <= 0 {
		return -1, errors.New("only positives allowed")
	}
	return v * v, nil
}
