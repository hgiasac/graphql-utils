package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

// Client abstracts the interface provided by hasura/go-graphql-client
// so their implementation can be replaced by something else
type Client interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}, options ...graphql.Option) error
	QueryRaw(ctx context.Context, q any, variables map[string]any, options ...graphql.Option) ([]byte, error)
	Mutate(ctx context.Context, m interface{}, variables map[string]interface{}, options ...graphql.Option) error
	MutateRaw(ctx context.Context, m interface{}, variables map[string]interface{}, options ...graphql.Option) ([]byte, error)
	Exec(ctx context.Context, query string, m any, variables map[string]any, options ...graphql.Option) error
	ExecRaw(ctx context.Context, query string, variables map[string]any, options ...graphql.Option) ([]byte, error)
}
