package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	t.Run("ok, alphabet", func(t *testing.T) {
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
	})

	t.Run("ok, underscore", func(t *testing.T) {
		// Given
		l := Lexer{
			input: "__hello",
		}
		l.Reset()

		// When
		output, err := l.parseName()

		// Then
		assert.NoError(t, err)
		assert.Equal(t, "__hello", output)
	})

	t.Run("fail: not _ or alphabet", func(t *testing.T) {
		// Given
		l := Lexer{
			input: "9hello",
		}
		l.Reset()

		// When
		_, err := l.parseName()

		// Then
		assert.Error(t, err)
	})
}
