package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSelection(t *testing.T) {
	t.Run("without parameter", func(t *testing.T) {
		l := Lexer{input: "SomeQuery"}
		l.Reset()
		l.push('\\')

		s, err := l.parseSelection()
		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", s.Name)
	})

	t.Run("with subselection", func(t *testing.T) {
		l := Lexer{input: `SomeQuery {
			subQuery	
		}`}
		l.Reset()
		l.push('\\')

		s, err := l.parseSelection()
		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", s.Name)
		assert.Equal(t, "subQuery", s.InnerSelection["subQuery"].Name)
	})

	t.Run("with arguments", func(t *testing.T) {
		l := Lexer{input: `SomeQuery(id: 123) {
			subQuery	
		}`}
		l.Reset()
		l.push('\\')

		s, err := l.parseSelection()
		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", s.Name)
		assert.Equal(t, "subQuery", s.InnerSelection["subQuery"].Name)
		assert.Equal(t, "id", s.Arguments["id"].Key)
	})
}

func TestParseSelectionSet(t *testing.T) {
	t.Run("with correct separator", func(t *testing.T) {
		l := Lexer{input: `{
		query1, query2
		query3
	}`}
		l.Reset()

		s, err := l.parseSelectionSet()

		assert.NoError(t, err)
		assert.Equal(t, "query1", s["query1"].Name)
		assert.Equal(t, "query2", s["query2"].Name)
		assert.Equal(t, "query3", s["query3"].Name)
	})

	t.Run("with incorrect separator", func(t *testing.T) {
		l := Lexer{input: `{
		query1 query2
		query3
	}`}
		l.Reset()

		_, err := l.parseSelectionSet()

		assert.Error(t, err)
	})

	t.Run("with nested value", func(t *testing.T) {
		l := Lexer{input: `{
		query1(id: 123) {
			query3	
		},
		query2
	}`}
		l.Reset()

		s, err := l.parseSelectionSet()

		assert.NoError(t, err)
		assert.Equal(t, "query1", s["query1"].Name)
		assert.Equal(t, "query3", s["query1"].InnerSelection["query3"].Name)
		assert.Equal(t, "id", s["query1"].Arguments["id"].Key)
		assert.Equal(t, "query2", s["query2"].Name)
	})

}
