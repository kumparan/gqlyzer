package gqlyzer

import (
	"testing"

	"github.com/kumparan/gqlyzer/token/operation"
	"github.com/stretchr/testify/assert"
)

func TestParseWithVariable(t *testing.T) {
	l := Lexer{input: `query SomeOperation {
			SomeQuery(id: $id) {
				subQuery
			}
		}`}
	l.Reset()
	s, err := l.ParseWithVariables(`
		{
			"id": "danu"
		}
	`)

	assert.NoError(t, err)
	assert.Equal(t, operation.Query, s.Type)
	assert.Equal(t, "SomeOperation", s.Name)
	assert.Equal(t, "SomeQuery", s.Selections["SomeQuery"].Name)
	assert.Equal(t, "id", s.Selections["SomeQuery"].Arguments["id"].Key)
	assert.Equal(t, `"danu"`, s.Selections["SomeQuery"].Arguments["id"].Value)
	assert.Equal(t, "subQuery", s.Selections["SomeQuery"].InnerSelection["subQuery"].Name)
}
