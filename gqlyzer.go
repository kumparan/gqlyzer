// Package gqlyzer graphql query lexical analyzer.
//
// The analysis is delegated to github.com/vektah/gqlparser/v2, a
// spec-compliant GraphQL parser. This package keeps gqlyzer's original public
// API and result types, so existing callers do not change.
package gqlyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/lexer"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/kumparan/gqlyzer/v2/token"
	"github.com/kumparan/gqlyzer/v2/token/operation"
)

// ErrEOF is retained for backward compatibility with callers that compare
// against it. The current implementation never returns it: an empty or
// whitespace-only document is not an error, and a malformed document returns a
// parse error that points at a line and column.
//
// Deprecated: nothing returns this value anymore.
var ErrEOF = errors.New("end of file")

// ErrOperationNotFound is returned when Options.OperationName names an
// operation the document does not define. Without it a mistyped name is
// indistinguishable from a document that selects nothing, which reads as
// harmless when it is not.
var ErrOperationNotFound = errors.New("operation not found")

// DefaultMaxTokenLimit bounds how many tokens a single document may contain.
// It guards against hostile input, since queries usually arrive from outside.
// Raise it through Options if a legitimate query is ever rejected.
const DefaultMaxTokenLimit = 15000

// defaultMaxSelectionNodes bounds how many selections one document may expand
// to. Fragments that reference each other can multiply out far beyond the size
// of the text, so the token limit alone is not enough of a guard.
const defaultMaxSelectionNodes = 50000

// keywordFragment is the keyword that opens a fragment definition.
const keywordFragment = "fragment"

// Options tunes parsing. The zero value is the recommended configuration.
type Options struct {
	// MaxTokenLimit caps the number of tokens in a document.
	// Zero means DefaultMaxTokenLimit. A negative value disables the cap.
	MaxTokenLimit int

	// OperationName selects which operation to return from a document that
	// holds more than one, the way a client's operationName does. Empty means
	// the first operation, which is what gqlyzer has always done.
	//
	// When the document defines no operation of that name, Parse and
	// ParseOperationType return ErrOperationNotFound.
	OperationName string

	// DisableFragmentExpansion stops fragment spreads and inline fragments
	// from being folded into the selection set of their parent.
	//
	// By default they are folded in, so that
	//
	//	{ edges { object { ... on Story { id title } } } }
	//
	// reports id and title under object. With expansion off, fragments are
	// skipped and object reports no fields at all.
	DisableFragmentExpansion bool

	// MaxSelectionNodes caps how many selections one document may expand to.
	// Zero means defaultMaxSelectionNodes. A negative value disables the cap.
	//
	// Reaching the cap truncates the selection set without reporting an
	// error, so do not use the result for an authorization decision unless
	// the cap is disabled.
	MaxSelectionNodes int
}

// Lexer analyzes a single GraphQL document.
//
// A Lexer is immutable once built. Its methods may be called more than once
// and from several goroutines at the same time; each call re-analyzes the
// document from the start and shares no state with any other call.
type Lexer struct {
	input string
	opts  Options
}

// New initializes a Lexer with the default options.
func New(gql string) (l *Lexer) {
	return &Lexer{input: gql}
}

// NewWithOptions initializes a Lexer with explicit options.
func NewWithOptions(gql string, opts Options) (l *Lexer) {
	return &Lexer{input: gql, opts: opts}
}

// Reset is retained for backward compatibility. The Lexer no longer carries
// parsing state between calls, so this does nothing.
func (l *Lexer) Reset() {}

// Parse analyzes the document and returns its operation.
//
// An empty or whitespace-only document yields a zero Operation and a nil
// error. A document that holds only fragment definitions does the same.
func (l *Lexer) Parse() (token.Operation, error) {
	return l.parse(nil)
}

// ParseOperationType returns the operation type without walking the body.
//
// With no Options.OperationName set it reads only the first significant
// token, so it stays cheap and does not care whether the rest of the document
// is well formed. A document with no operation in it — empty, whitespace
// only, or holding nothing but fragment definitions — yields an empty type
// and a nil error, matching Parse.
//
// When Options.OperationName is set the document has to be parsed to find
// that operation, so this costs the same as Parse and reports the same
// errors, including ErrOperationNotFound.
func (l *Lexer) ParseOperationType() (operation.Type, error) {
	if l.opts.OperationName != "" {
		doc, err := l.parseDocument()
		if err != nil {
			return "", err
		}

		op, err := l.pickOperation(doc)
		if err != nil || op == nil {
			return "", err
		}

		return operationType(op.Operation), nil
	}

	lex := lexer.New(&ast.Source{Input: l.input})

	for {
		tok, err := lex.ReadToken()
		if err != nil {
			return "", err
		}

		switch tok.Kind { //nolint:exhaustive
		case lexer.Comment:
			continue // a leading comment says nothing about the operation
		case lexer.EOF:
			// An empty or whitespace-only document is not an error.
			return "", nil
		case lexer.BraceL:
			// An anonymous operation is a query.
			return operation.Query, nil
		case lexer.Name:
			switch tok.Value {
			case string(ast.Query):
				return operation.Query, nil
			case string(ast.Mutation):
				return operation.Mutation, nil
			case string(ast.Subscription):
				return operation.Subscription, nil
			case keywordFragment:
				// A document that leads with a fragment definition declares no
				// operation. Parse reports that as a zero value and a nil
				// error; report it the same way here.
				return "", nil
			}
		}

		return "", fmt.Errorf("unknown definition: %s", tok.Value)
	}
}

// ParseWithVariables analyzes the document and resolves variable references
// using variables, a JSON object such as {"id": "123", "size": 10}.
//
// A variable that the JSON object does not mention keeps its reference form in
// the result, for example "$id".
func (l *Lexer) ParseWithVariables(variables string) (token.Operation, error) {
	vars := make(map[string]any)
	if strings.TrimSpace(variables) != "" {
		if err := json.Unmarshal([]byte(variables), &vars); err != nil {
			return token.Operation{}, err
		}
	}

	return l.parse(vars)
}

func (l *Lexer) parse(vars map[string]any) (token.Operation, error) {
	doc, err := l.parseDocument()
	if err != nil {
		return token.Operation{}, err
	}

	op, err := l.pickOperation(doc)
	if err != nil {
		return token.Operation{}, err
	}
	if op == nil {
		return token.Operation{}, nil
	}

	result := token.Operation{
		Type: operationType(op.Operation),
		Name: op.Name,
	}

	for _, vd := range op.VariableDefinitions {
		result.Variables = append(result.Variables, token.Parameter{Name: vd.Variable})
	}

	budget := l.opts.MaxSelectionNodes
	if budget == 0 {
		budget = defaultMaxSelectionNodes
	}

	w := &walker{
		fragments:      doc.Fragments,
		vars:           vars,
		unresolved:     map[string]bool{},
		expandFragment: !l.opts.DisableFragmentExpansion,
		budget:         budget,
		unlimited:      budget < 0,
	}
	result.Selections = w.selectionSet(op.SelectionSet, map[string]bool{})

	if len(w.unresolved) > 0 {
		result.UnresolvedFragments = make([]string, 0, len(w.unresolved))
		for name := range w.unresolved {
			result.UnresolvedFragments = append(result.UnresolvedFragments, name)
		}
		sort.Strings(result.UnresolvedFragments)
	}

	return result, nil
}

func (l *Lexer) parseDocument() (*ast.QueryDocument, error) {
	limit := l.opts.MaxTokenLimit
	switch {
	case limit == 0:
		limit = DefaultMaxTokenLimit
	case limit < 0:
		limit = 0 // gqlparser reads 0 as unlimited
	}

	doc, err := parser.ParseQueryWithTokenLimit(&ast.Source{Input: l.input}, limit)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// pickOperation returns the operation the options ask for. A nil operation
// with a nil error means the document declares none, which is not a failure.
func (l *Lexer) pickOperation(doc *ast.QueryDocument) (*ast.OperationDefinition, error) {
	if doc == nil || len(doc.Operations) == 0 {
		if l.opts.OperationName != "" {
			return nil, fmt.Errorf("%w: %q", ErrOperationNotFound, l.opts.OperationName)
		}

		return nil, nil
	}

	if l.opts.OperationName == "" {
		return doc.Operations[0], nil
	}

	op := doc.Operations.ForName(l.opts.OperationName)
	if op == nil {
		return nil, fmt.Errorf("%w: %q", ErrOperationNotFound, l.opts.OperationName)
	}

	return op, nil
}

func operationType(op ast.Operation) operation.Type {
	switch op {
	case ast.Mutation:
		return operation.Mutation
	case ast.Subscription:
		return operation.Subscription
	case ast.Query:
		return operation.Query
	default:
		return operation.Type(strings.ToUpper(string(op)))
	}
}

// walker turns a gqlparser selection tree into a token.SelectionSet.
type walker struct {
	fragments      ast.FragmentDefinitionList
	vars           map[string]any
	unresolved     map[string]bool
	expandFragment bool
	budget         int
	unlimited      bool
}

// selectionSet flattens sels into a token.SelectionSet.
//
// activeFragments holds the fragment names on the current expansion path, so
// that a fragment cycle terminates instead of recursing forever.
func (w *walker) selectionSet(sels ast.SelectionSet, activeFragments map[string]bool) token.SelectionSet {
	set := make(token.SelectionSet)
	w.collect(set, sels, activeFragments)

	return set
}

func (w *walker) collect(set token.SelectionSet, sels ast.SelectionSet, activeFragments map[string]bool) {
	for _, sel := range sels {
		if !w.spend() {
			return
		}

		switch s := sel.(type) {
		case *ast.Field:
			w.mergeField(set, s, activeFragments)

		case *ast.InlineFragment:
			if !w.expandFragment {
				continue
			}
			// "... on Story { id }" contributes id to the parent.
			w.collect(set, s.SelectionSet, activeFragments)

		case *ast.FragmentSpread:
			if !w.expandFragment || activeFragments[s.Name] {
				continue
			}
			def := w.fragments.ForName(s.Name)
			if def == nil {
				// The document spreads a fragment it does not define, which
				// the specification's "fragment spread target defined" rule
				// forbids. A fragment name is not a field, so it must not go
				// into the selection set; report it separately instead of
				// dropping it silently.
				w.unresolved[s.Name] = true
				continue
			}

			activeFragments[s.Name] = true
			w.collect(set, def.SelectionSet, activeFragments)
			delete(activeFragments, s.Name)
		}
	}
}

func (w *walker) mergeField(set token.SelectionSet, f *ast.Field, activeFragments map[string]bool) {
	sel := token.Selection{
		Name:  f.Name,
		Alias: f.Alias,
	}
	// gqlparser always fills Alias, even when the query gave none. gqlyzer
	// has always left Alias empty in that case.
	if sel.Alias == sel.Name {
		sel.Alias = ""
	}

	if len(f.Arguments) > 0 {
		sel.Arguments = w.argumentSet(f.Arguments)
	}

	if len(f.SelectionSet) > 0 {
		sel.InnerSelection = w.selectionSet(f.SelectionSet, activeFragments)
	}

	w.mergeSelection(set, sel)
}

// mergeSelection adds sel to set, combining it with an existing entry of the
// same name rather than replacing it. Two branches of a query can ask for the
// same field, as in "... on User { id }" beside "... on Publisher { id }".
func (w *walker) mergeSelection(set token.SelectionSet, sel token.Selection) {
	existing, found := set[sel.Name]
	if !found {
		set[sel.Name] = sel
		return
	}

	if existing.Alias == "" {
		existing.Alias = sel.Alias
	}

	if existing.Arguments == nil {
		existing.Arguments = sel.Arguments
	} else {
		for k, v := range sel.Arguments {
			if _, dup := existing.Arguments[k]; !dup {
				existing.Arguments[k] = v
			}
		}
	}

	if existing.InnerSelection == nil {
		existing.InnerSelection = sel.InnerSelection
	} else {
		for k, v := range sel.InnerSelection {
			if _, dup := existing.InnerSelection[k]; !dup {
				existing.InnerSelection[k] = v
			}
		}
	}

	set[sel.Name] = existing
}

// spend draws one unit from the expansion budget and reports whether work may
// continue.
func (w *walker) spend() bool {
	if w.unlimited {
		return true
	}
	if w.budget <= 0 {
		return false
	}
	w.budget--

	return true
}

func (w *walker) argumentSet(args ast.ArgumentList) token.ArgumentSet {
	set := make(token.ArgumentSet, len(args))
	for _, a := range args {
		set[a.Name] = w.argument(a.Name, a.Value)
	}

	return set
}

func (w *walker) argument(name string, v *ast.Value) token.Argument {
	arg := token.Argument{Key: name}

	if v != nil && v.Kind == ast.ObjectValue {
		// An object argument is reported through ObjectValue, as before.
		arg.ObjectValue = make(token.ArgumentSet, len(v.Children))
		for _, child := range v.Children {
			arg.ObjectValue[child.Name] = w.argument(child.Name, child.Value)
		}

		return arg
	}

	arg.Value = w.renderValue(v)

	return arg
}

// renderValue formats an argument value the way gqlyzer has always reported
// it: strings keep their quotes, everything else keeps its literal text.
func (w *walker) renderValue(v *ast.Value) string {
	if v == nil {
		return ""
	}

	switch v.Kind {
	case ast.Variable:
		if resolved, found := w.vars[v.Raw]; found {
			return renderJSON(resolved)
		}
		if v.VariableDefinition != nil && v.VariableDefinition.DefaultValue != nil {
			return w.renderValue(v.VariableDefinition.DefaultValue)
		}

		return "$" + v.Raw

	case ast.StringValue, ast.BlockValue:
		return strconv.Quote(v.Raw)

	case ast.IntValue, ast.FloatValue, ast.BooleanValue, ast.NullValue, ast.EnumValue:
		return v.Raw

	case ast.ListValue:
		parts := make([]string, 0, len(v.Children))
		for _, child := range v.Children {
			parts = append(parts, w.renderValue(child.Value))
		}

		return "[" + strings.Join(parts, ", ") + "]"

	case ast.ObjectValue:
		parts := make([]string, 0, len(v.Children))
		for _, child := range v.Children {
			parts = append(parts, child.Name+": "+w.renderValue(child.Value))
		}

		return "{" + strings.Join(parts, ", ") + "}"

	default:
		return v.Raw
	}
}

// renderJSON formats a value decoded from the variables JSON.
func renderJSON(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return strconv.Quote(t)
	case bool:
		return strconv.FormatBool(t)
	case float64:
		// encoding/json decodes every number as float64. Print whole numbers
		// without a trailing ".0" so that an ID reads as 19, not 19.0.
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}

		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}

		return string(encoded)
	}
}
