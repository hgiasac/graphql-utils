package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// client represents a new custom graphql client
type client struct {
	*graphql.Client
	logger zerolog.Logger
}

// NewClient creates a new custom graphql client
func NewClient(url string, httpClient graphql.Doer) *client {
	return &client{
		Client: graphql.NewClient(url, httpClient),
		logger: log.Level(zerolog.GlobalLevel()),
	}
}

// WithLogger creates a new client with the input logger
func (c *client) WithLogger(logger zerolog.Logger) *client {
	return &client{
		Client: c.Client,
		logger: logger,
	}
}

// WithDebug set debug mode
func (c *client) WithDebug(debug bool) *client {
	l := c.logger
	if debug {
		l = l.Level(zerolog.DebugLevel)
	}
	return &client{
		Client: c.Client.WithDebug(debug),
		logger: l,
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
	logEvent := c.logger.Debug()
	isDebug := c.logger.GetLevel() <= zerolog.DebugLevel
	if isDebug {
		logEvent = logEvent.Interface("variables", variables).Interface("query", query)
	}
	err := c.Client.Exec(ctx, query, q, variables, options...)

	if isDebug {
		if err != nil {
			logEvent.Err(err).Msg(msg)
		} else {
			logEvent.Interface("response", q).Msg(msg)
		}
	}
	return err
}

func (c *client) execRaw(ctx context.Context, query string, variables map[string]any, msg string, options ...graphql.Option) ([]byte, error) {
	logEvent := c.logger.Debug()
	isDebug := c.logger.GetLevel() <= zerolog.DebugLevel
	if isDebug {
		logEvent = logEvent.Interface("variables", variables).Interface("query", query)
	}
	bs, err := c.Client.ExecRaw(ctx, query, variables, options...)

	if isDebug {
		if err != nil {
			logEvent.Err(err).Msg(msg)
		} else {
			logEvent.RawJSON("response", bs).Msg(msg)
		}
	}
	return bs, err
}
