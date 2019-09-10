package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	// Given
	l := Lexer{
		input: "hello",
	}
	l.Reset()

	// When
	output := l.parseName()

	// Then
	assert.Equal(t, "hello", output)
}
