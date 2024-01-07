package test

import (
	"context"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

func TestNewMockGraphQLClientQueries(t *testing.T) {

	client := NewMockGraphQLClientQueries(map[string]string{
		`{user{id}}`: EncodeMockGraphQLResponse(map[string]any{
			"user": map[string]any{
				"id": 1,
			},
		}, nil),
		"query GetUser($foo:String!){user{id}}": EncodeMockGraphQLResponse(map[string]any{
			"user": map[string]any{
				"id": 2,
			},
		}, nil),
	})

	var userQuery struct {
		User struct {
			ID int
		}
	}

	assert.NoError(t, client.Query(context.TODO(), &userQuery, nil))
	assert.Equal(t, 1, userQuery.User.ID)

	assert.NoError(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetUser")))
	assert.Equal(t, 2, userQuery.User.ID)

	assert.ErrorContains(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetNotExistUser")), "validation-failed")
}

func TestNewMockGraphQLClientSingle(t *testing.T) {
	var userQuery struct {
		User struct {
			ID int
		}
	}
	clientStr := NewMockGraphQLClientSingle(`{"data": {"user": {"id": 1}}}`, nil)
	assert.NoError(t, clientStr.Query(context.TODO(), &userQuery, nil))
	assert.Equal(t, 1, userQuery.User.ID)

	clientObj := NewMockGraphQLClientSingle(map[string]any{
		"user": map[string]any{
			"id": 2,
		},
	}, nil)
	assert.NoError(t, clientObj.Query(context.TODO(), &userQuery, nil))
	assert.Equal(t, 2, userQuery.User.ID)
}

func TestNewMockGraphQLAffectedRowsResponse(t *testing.T) {
	var userQuery struct {
		InsertUser struct {
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"insert_user(objects: $objects)"`
	}
	clientStr := NewMockGraphQLClientSingle(NewMockGraphQLAffectedRowsResponse("insert_user", 1), nil)
	assert.NoError(t, clientStr.Query(context.TODO(), &userQuery, nil))
	assert.Equal(t, 1, userQuery.InsertUser.AffectedRows)
}

func TestNewMockGraphQLClient(t *testing.T) {

	client := NewMockGraphQLClient([]MockGraphQLResponse{
		{
			Request: graphql.GraphQLRequestPayload{
				Query: "{user{id}}",
			},
			Response: map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"id": 1,
					},
				},
			},
		},
		{
			Request: graphql.GraphQLRequestPayload{
				Query: "query GetUser($foo:String!){user{id}}",
				Variables: map[string]interface{}{
					"foo": "bar",
				},
			},
			Response: map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"id": 2,
					},
				},
			},
		},
	})

	var userQuery struct {
		User struct {
			ID int
		}
	}

	assert.NoError(t, client.Query(context.TODO(), &userQuery, nil))
	assert.Equal(t, 1, userQuery.User.ID)

	assert.ErrorContains(t, client.Query(context.TODO(), &userQuery, nil, graphql.OperationName("GetUser")), "validation-failed")
	assert.Equal(t, 1, userQuery.User.ID)

	assert.NoError(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetUser")))
	assert.Equal(t, 2, userQuery.User.ID)

	assert.ErrorContains(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetNotExistUser")), "validation-failed")
}
