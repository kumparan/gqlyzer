package gqlyzer

import "github.com/stretchr/testify/assert"
import "testing"

func TestParseMutationKeyword(t *testing.T) {
	t.Run(`should return no error when given correct keyword`, func(t *testing.T) {
		l := Lexer{input: "mutation"}
		l.Reset()

		err := l.parseKeyword("mutation")

		assert.NoError(t, err)
	})

	t.Run(`should return error when given mismatch keyword`, func(t *testing.T) {
		l := Lexer{input: "mutation"}
		l.Reset()

		err := l.parseKeyword("query")

		assert.Error(t, err)
	})
}
