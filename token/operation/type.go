package operation

// Type is the operation type
type Type string

const (
	// Mutation type
	Mutation Type = "MUTATION"
	// Query type
	Query Type = "QUERY"
	// Subscription type
	Subscription Type = "SUBSCRIPTION"
)
