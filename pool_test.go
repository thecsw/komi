package komi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoolCreationClosing(t *testing.T) {
	simplePool := NewWithSettings(WorkSimple(squareSimple), &Settings{
		Laborers: 10,
		Debug:    true,
		Name:     "Simple Pool",
	})
	assert.NotNil(t, simplePool, "simple pool")
	simplePool.Close()
	assert.Equal(t, true, simplePool.closed, "closed flag")
	assert.NotEmpty(t, simplePool.closedSignal, "closed signal channel")

	simplePoolWithErrors := NewWithSettings(WorkSimpleWithErrors(squareSimpleWithErrors), &Settings{
		Debug: true,
		Name:  "Simple Pool With Errors",
	})
	assert.NotNil(t, simplePoolWithErrors, "simple pool with errors")
	defer simplePoolWithErrors.Close()

	regularPool := New(Work(squareRegular))
	regularPool.Debug()
	assert.NotNil(t, regularPool, "regular pool")
	defer regularPool.Close()

	regularPoolWithErrors := NewWithSettings(WorkWithErrors(squarReguralWithErrors), &Settings{
		Debug: true,
		Name:  "Regular Pool With Errors",
	})
	assert.NotNil(t, regularPoolWithErrors, "regular pool with errors")
	defer regularPoolWithErrors.Close()
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
