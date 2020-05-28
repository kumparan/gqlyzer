package token

import "github.com/kumparan/gqlyzer/token/operation"

type (
	// SelectionSet is list of selection
	SelectionSet map[string]Selection
	// Variables given to the operation
	Variables map[string]string
)

// Parameter containing information of a query or operation parameter
// including name and argument
// TODO: wont be implemented for now
type Parameter struct {
	Name string
}

// Operation use to contain information of an operation
type Operation struct {
	Type operation.Type
	Name string
	// TODO: wont be implemented for now
	Variables  []Parameter
	Selections SelectionSet
}

// Selection containing information of a selection
// including query, mutation or field recursively
type Selection struct {
	// Name field
	Name string
	// will be empty, if the selection have no sub field
	InnerSelection SelectionSet
	// Arguments in the selection
	Arguments ArgumentSet
	// TODO: wont be implemented for now
	Alias string
}

// ArgumentSet set of arguments
type ArgumentSet map[string]Argument

// Argument operation argument
type Argument struct {
	Key         string
	Value       string
	ObjectValue ArgumentSet
}
