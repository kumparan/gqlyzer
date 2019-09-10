package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseName_Ok(t *testing.T) {
	// Given
	l := Lexer{
		input: "hello",
	}
	l.Reset()

	// When
	output, err := l.parseName()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "hello", output)
}

func TestParseName_Fail(t *testing.T) {
	// Given
	l := Lexer{
		input: "9hello",
	}
	l.Reset()

	// When
	_, err := l.parseName()

	// Then
	assert.Error(t, err)
}
