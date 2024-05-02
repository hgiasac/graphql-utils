package test

import (
	"context"
	"strings"
	"testing"

	"github.com/hasura/go-graphql-client"
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

	assertNoError(t, client.Query(context.TODO(), &userQuery, nil))
	assertEqual(t, 1, userQuery.User.ID)

	assertNoError(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetUser")))
	assertEqual(t, 2, userQuery.User.ID)

	assertError(t, client.Query(context.TODO(), &userQuery, map[string]any{
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
	assertNoError(t, clientStr.Query(context.TODO(), &userQuery, nil))
	assertEqual(t, 1, userQuery.User.ID)

	clientObj := NewMockGraphQLClientSingle(map[string]any{
		"user": map[string]any{
			"id": 2,
		},
	}, nil)
	assertNoError(t, clientObj.Query(context.TODO(), &userQuery, nil))
	assertEqual(t, 2, userQuery.User.ID)
}

func TestNewMockGraphQLAffectedRowsResponse(t *testing.T) {
	var userQuery struct {
		InsertUser struct {
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"insert_user(objects: $objects)"`
	}
	clientStr := NewMockGraphQLClientSingle(NewMockGraphQLAffectedRowsResponse("insert_user", 1), nil)
	assertNoError(t, clientStr.Query(context.TODO(), &userQuery, nil))
	assertEqual(t, 1, userQuery.InsertUser.AffectedRows)
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

	assertNoError(t, client.Query(context.TODO(), &userQuery, nil))
	assertEqual(t, 1, userQuery.User.ID)

	assertError(t, client.Query(context.TODO(), &userQuery, nil, graphql.OperationName("GetUser")), "validation-failed")
	assertEqual(t, 1, userQuery.User.ID)

	assertNoError(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetUser")))
	assertEqual(t, 2, userQuery.User.ID)

	assertError(t, client.Query(context.TODO(), &userQuery, map[string]any{
		"foo": "bar",
	}, graphql.OperationName("GetNotExistUser")), "validation-failed")
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func assertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Error("expected error, got: nil")
	}

	if !strings.Contains(err.Error(), message) {
		t.Errorf("expected error content: %v, got: %s", err, message)
	}
}

func assertEqual(t *testing.T, expected any, got any) {
	if !deepEqual(expected, got) {
		t.Errorf("not equal, expected: %+v, got: %+v", expected, got)
	}
}
