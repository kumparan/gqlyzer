package gqlyzer

import (
	"testing"

	"github.com/kumparan/gqlyzer/token/operation"
	"github.com/stretchr/testify/assert"
)

func TestParseOperation(t *testing.T) {
	t.Run("with anonymous operation", func(t *testing.T) {
		l := Lexer{input: `{
			SomeQuery(id: 123) {
				subQuery	
			}
		}`}
		l.Reset()
		l.push('\\')

		s, err := l.parseOperation()
		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "", s.Name)
		assert.Equal(t, "SomeQuery", s.Selections["SomeQuery"].Name)
		assert.Equal(t, "id", s.Selections["SomeQuery"].Arguments["id"].Key)
		assert.Equal(t, "123", s.Selections["SomeQuery"].Arguments["id"].Value)
		assert.Equal(t, "subQuery", s.Selections["SomeQuery"].InnerSelection["subQuery"].Name)
	})
}
