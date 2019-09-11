package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSelection(t *testing.T) {
	t.Run("without parameter", func(t *testing.T) {
		l := Lexer{input: "SomeQuery"}
		l.Reset()

		_, err := l.parseSelection()
		assert.NoError(t, err)
	})
}

func TestParseSelectionSet(t *testing.T) {
	l := Lexer{input: `{
		query1
		query2
	}`}
	l.Reset()

	_, err := l.parseSelectionSet()

	assert.NoError(t, err)
}
