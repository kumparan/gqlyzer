# GQLyzer

[![Go Reference](https://pkg.go.dev/badge/github.com/kumparan/gqlyzer/v2.svg)](https://pkg.go.dev/github.com/kumparan/gqlyzer/v2)

GQLyzer reports what an incoming GraphQL document asks for: its operation
type, its operation name, and the fields it selects, down the tree, with
their arguments.

It answers questions you want to answer before the request reaches a
resolver — is this a query or a mutation, which operation is this, which
fields does it touch — without needing a schema and without executing
anything.

```go
op, err := gqlyzer.New(`query FindStory($slug: String!) {
    FindStoryBySlug(slug: $slug) {
        id
        publisher { slug }
    }
}`).Parse()

op.Type                                              // "QUERY"
op.Name                                              // "FindStory"
op.Selections["FindStoryBySlug"].Arguments["slug"]   // {Key: "slug", Value: "$slug"}
op.Selections["FindStoryBySlug"].
    InnerSelection["publisher"].
    InnerSelection["slug"].Name                      // "slug"
```

## Installation

```bash
go get github.com/kumparan/gqlyzer/v2
```

```go
import "github.com/kumparan/gqlyzer/v2"
```

v2 is a different module path from v1, so the two can sit side by side
while you migrate.

## Why v2

v1 shipped its own hand-written lexer. v2 takes a different approach: it
delegates the parsing to [`vektah/gqlparser/v2`][gqlparser] and maps the
result onto the same `token.Operation` shape v1 returned.

The reason for the change was a sample of 211 real queries taken from
production. **All 211 failed to parse on v1**, and all 211 are valid
GraphQL — confirmed against two independent reference parsers. The
failures fell into three groups, and they shared one root cause: v1 had no
tokenizer, so it treated a field boundary as "a newline or a comma"
instead of as a token boundary.

| v1 error | Cause |
| --- | --- |
| `expected separator, but got: #` | Anything written on one line. Even `query { a, b }` did not parse. |
| `end of file` | An inline sub-selection such as `publisher { slug }` swallowed its closing brace, so every later brace was off by one. |
| `first character of an identifier...` | An escaped quote inside a string value (`"{\"k\":\"v\"}"`) ended the string early, breaking every mutation carrying a JSON payload. |

Several other v1 bugs returned a *wrong answer with no error*, which is
worse than failing:

- `parseArray` advanced its cursor twice, so an argument list followed by
  another argument silently lost the arguments and promoted a sub-field to
  the top level.
- Fragments were reported as fields. `... on Story { id }` reported a field
  named `Story`; `...ItemDetails` reported a field named `ItemDetails`.
- `ParseWithVariables` substituted variables with `strings.ReplaceAll`, so
  `$id` also matched the start of `$idType` and Go's map iteration order
  decided which won. The same input gave different answers on different
  runs.
- A `)` inside a variable's default value ended the variable block early
  and emptied the selection set, with no error.
- `read()` returned `rune(input[cursor])`, a single byte, so any non-ASCII
  byte was mangled.

Rather than patch these one by one, v2 hands the job to a parser that
already implements the specification and is exercised far more widely than
this package could be.

## What v2 changes

The public API is unchanged: `New`, `Reset`, `Parse`, `ParseOperationType`
and `ParseWithVariables` keep their signatures, and the result types in
`token` are the same. Most callers compile and keep working.

What changes is the answers you get back.

### Documents that used to fail now parse

One-line queries, inline sub-selections, escaped quotes and block strings,
comments, nested and empty lists, `... Frag` written with a space, and
variable defaults containing punctuation. If your code has a fallback path
for "gqlyzer could not read this", expect it to go quiet.

### Fragments now report fields, not fragment names

This is the change most likely to affect you.

```graphql
{ edges { object { ... on Story { id title } } } }
```

| | `Selections["edges"].InnerSelection["object"].InnerSelection` |
| --- | --- |
| v1 | `{"Story": ...}` — the type condition, as if it were a field |
| v2 | `{"id": ..., "title": ...}` — the fields actually requested |

Inline fragments and named fragment spreads are folded into their parent
selection set, so the result reflects the fields the document asks for.
Where two branches request the same field, the entries are merged rather
than overwriting each other. Set `Options.DisableFragmentExpansion` to skip
fragments instead.

### Errors are no longer swallowed

v1 discarded errors from arguments and sub-selections, and often returned a
partial result with a `nil` error. v2 returns the parse error, with a line
and column. Callers that only checked `err != nil` will now see failures
they previously did not.

### Value formatting

Strings, integers, floats, booleans, enums, `null` and object arguments are
formatted exactly as v1 formatted them: `"hello"` keeps its quotes, `USER`
and `12` keep their literal text, and an object argument is reported
through `Argument.ObjectValue` with `Argument.Value` left empty.

List arguments render as `[a, b]`, with a space after each comma. v1 rarely
parsed a list at all — where it did, it dropped the argument — so there is
little v1 output to be compatible with.

### `ErrEOF` is deprecated

It is still exported and its message is still `"end of file"`, so code
comparing against it still compiles. Nothing returns it any more: an empty
or whitespace-only document is not an error, and a malformed one returns a
real parse error.

### Other notes

- `Reset` is a no-op. The `Lexer` carries no parse state, so `Parse` may be
  called more than once on the same value.
- `Selections` is still keyed by field name, not by alias, so two aliases of
  the same field still collapse into one entry.
- Identifiers follow the specification: `[_A-Za-z][_0-9A-Za-z]*`. Non-ASCII
  names are rejected rather than mangled.

## API

### Parsing

```go
l := gqlyzer.New(document)

op, err := l.Parse()                          // full analysis
ot, err := l.ParseOperationType()             // operation type only
op, err := l.ParseWithVariables(variablesJSON) // resolve $variables
```

`ParseOperationType` reads only the first significant token, so it stays
cheap and does not care whether the rest of the document is well formed.
Use it when routing on query/mutation/subscription is all you need.

`ParseWithVariables` takes the variables as a JSON object and resolves
references in argument values. A variable the JSON does not mention keeps
its reference form, for example `"$id"`.

```go
op, _ := gqlyzer.New(`query { a(size: $size) }`).
    ParseWithVariables(`{"size": 25}`)

op.Selections["a"].Arguments["size"].Value // "25"
```

An empty or whitespace-only document, or one holding only fragment
definitions, returns a zero `token.Operation` and a `nil` error.

### Options

```go
l := gqlyzer.NewWithOptions(document, gqlyzer.Options{
    MaxTokenLimit:            15000,
    OperationName:            "FindStory",
    DisableFragmentExpansion: false,
    MaxSelectionNodes:        50000,
})
```

| Option | Default | Purpose |
| --- | --- | --- |
| `MaxTokenLimit` | 15000 | Caps document size. Negative disables. |
| `OperationName` | first operation | Selects one operation from a multi-operation document, the way a client's `operationName` does. |
| `DisableFragmentExpansion` | `false` | Report fragments as skipped instead of folding them into the parent. |
| `MaxSelectionNodes` | 50000 | Caps how far fragments may expand. Negative disables. |

Both caps exist because documents usually arrive from outside. Fragments
that reference each other multiply out far beyond the size of the text: a
40-level fragment diamond is under 1.5 KB and expands to 2⁴⁰ selections
without a bound.

> **Note:** reaching `MaxSelectionNodes` truncates the selection set
> *without* reporting an error. Do not use the result for an authorization
> decision unless you disable the cap.

### Result types

```go
type Operation struct {
    Type       operation.Type // "QUERY" | "MUTATION" | "SUBSCRIPTION"
    Name       string
    Variables  []Parameter
    Selections SelectionSet   // map[string]Selection, keyed by field name
}

type Selection struct {
    Name           string
    Alias          string       // empty when the document gave none
    Arguments      ArgumentSet  // map[string]Argument
    InnerSelection SelectionSet
}

type Argument struct {
    Key         string
    Value       string      // formatted literal; empty for object arguments
    ObjectValue ArgumentSet // set when the value is an object
}
```

## Testing

```bash
make test        # lint + tests
make test-only   # tests only
```

The suite carries every test from v1, so the behavioural changes above are
visible as assertions rather than as surprises. Six v1 tests are gone; each
asserted something about the old lexer's own cursor rather than about the
analysis it produced, and the test file lists them and says what replaced
them.

`TestParse_ProductionCorpus` replays the 211 captured queries from
`testdata/queries.csv`. That file is **not** in version control, because
this repository is public and the capture carries real article text and
real author, publisher and channel IDs. The test skips when the file is
absent. See `.gitignore`.

## Limitations

- No schema awareness and no validation. GQLyzer reports what a document
  *asks for*; it cannot tell you whether those fields exist or whether the
  types line up. Pair it with a validator if you need that.
- `Selections` is a map keyed by field name, so a document selecting the
  same field twice under different aliases reports one entry.
- Directives are parsed but not reported.

## License

See [LICENSE](LICENSE).

[gqlparser]: https://github.com/vektah/gqlparser