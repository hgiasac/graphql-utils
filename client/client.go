package client

import (
	"context"
	"log/slog"

	"github.com/hasura/go-graphql-client"
)

// client represents a new custom graphql client
type client struct {
	*graphql.Client
	logger *slog.Logger
}

// NewClient creates a new custom graphql client
func NewClient(url string, httpClient graphql.Doer) *client {
	return &client{
		Client: graphql.NewClient(url, httpClient),
		logger: slog.Default(),
	}
}

// WithLogger creates a new client with the input logger
func (c *client) WithLogger(logger *slog.Logger) *client {
	return &client{
		Client: c.Client,
		logger: logger,
	}
}

// Query executes a graphql query request
func (c *client) Query(ctx context.Context, q any, variables map[string]any, options ...graphql.Option) error {
	query, err := graphql.ConstructQuery(q, variables, options...)
	if err != nil {
		return err
	}
	return c.exec(ctx, query, q, variables, "Query", options...)
}

// Query executes a graphql query request, return raw bytes
func (c *client) QueryRaw(ctx context.Context, q any, variables map[string]any, options ...graphql.Option) ([]byte, error) {

	query, err := graphql.ConstructQuery(q, variables, options...)
	if err != nil {
		return nil, err
	}
	return c.execRaw(ctx, query, variables, "QueryRaw", options...)
}

// Query executes a graphql mutation request
func (c *client) Mutate(ctx context.Context, m any, variables map[string]any, options ...graphql.Option) error {
	query, err := graphql.ConstructMutation(m, variables, options...)
	if err != nil {
		return err
	}
	return c.exec(ctx, query, m, variables, "Mutate", options...)
}

// Query executes a graphql mutation request, return raw bytes
func (c *client) MutateRaw(ctx context.Context, m any, variables map[string]any, options ...graphql.Option) ([]byte, error) {
	query, err := graphql.ConstructMutation(m, variables, options...)
	if err != nil {
		return nil, err
	}
	return c.execRaw(ctx, query, variables, "MutateRaw", options...)
}

// Exec executes a graphql request from raw query string
func (c *client) Exec(ctx context.Context, query string, m any, variables map[string]any, options ...graphql.Option) error {
	return c.exec(ctx, query, m, variables, "Exec", options...)
}

// Exec executes a graphql request from raw query string, return raw bytes
func (c *client) ExecRaw(ctx context.Context, query string, variables map[string]any, options ...graphql.Option) ([]byte, error) {
	return c.execRaw(ctx, query, variables, "ExecRaw", options...)
}

func (c *client) exec(ctx context.Context, query string, q any, variables map[string]any, msg string, options ...graphql.Option) error {
	var logAttrs []any
	isDebug := c.logger.Enabled(ctx, slog.LevelDebug)
	if isDebug {
		logAttrs = append(logAttrs, slog.Any("variables", variables), slog.String("query", query))
	}
	err := c.Client.Exec(ctx, query, q, variables, options...)

	if isDebug {
		if err != nil {
			logAttrs = append(logAttrs, slog.Any("error", err))
			c.logger.Error(msg, logAttrs...)
		} else {
			logAttrs = append(logAttrs, slog.Any("response", q))
			c.logger.Error(msg, logAttrs...)
		}
	}
	return err
}

func (c *client) execRaw(ctx context.Context, query string, variables map[string]any, msg string, options ...graphql.Option) ([]byte, error) {
	var logAttrs []any
	isDebug := c.logger.Enabled(ctx, slog.LevelDebug)
	if isDebug {
		logAttrs = append(logAttrs, slog.Any("variables", variables), slog.String("query", query))
	}
	bs, err := c.Client.ExecRaw(ctx, query, variables, options...)

	if isDebug {
		if err != nil {
			logAttrs = append(logAttrs, slog.Any("error", err))
			c.logger.Error(msg, logAttrs...)
		} else {
			logAttrs = append(logAttrs, slog.String("response", string(bs)))
			c.logger.Error(msg, logAttrs...)
		}
	}
	return bs, err
}
