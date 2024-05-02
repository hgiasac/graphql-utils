package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type MockGraphQLResponse struct {
	Request  graphql.GraphQLRequestPayload
	Response any
}

type MockDoType func(req *http.Request) (*http.Response, error)

type MockHTTPClient struct {
	MockDo MockDoType
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

// NewMockHTTPClient creates a mock client of type HTTPClient where the mock Do()
// implementation returns the reponse and status code provided as params to the function.
func NewMockHTTPClient(jsonResp string, statusCode int) *MockHTTPClient {

	return &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			r := io.NopCloser(bytes.NewReader([]byte(jsonResp)))
			return &http.Response{
				StatusCode: statusCode,
				Body:       r,
			}, nil
		},
	}
}

// NewMockGraphQLClient creates a mock graphql client of type HTTPClient where the mock Do()
// implementation returns the response and status code provided as params to the func.
func NewMockGraphQLClient(responses []MockGraphQLResponse) *graphql.Client {
	preparedQueries := make([]MockGraphQLResponse, 0, len(responses))
	for _, q := range responses {
		q.Request.Query = strings.TrimSpace(q.Request.Query)
		q.Request.OperationName = strings.TrimSpace(q.Request.OperationName)

		preparedQueries = append(preparedQueries, q)
	}
	return graphql.NewClient("/v1/graphql", &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return responseGraphQLError(err, []byte("io.ReadAll")), nil
			}

			var reqBody graphql.GraphQLRequestPayload
			err = json.Unmarshal(bodyBytes, &reqBody)

			if err != nil {
				return responseGraphQLError(err, bodyBytes), nil
			}

			var jsonResp []byte

			for _, resp := range responses {
				if resp.Request.Query != strings.TrimSpace(reqBody.Query) {
					continue
				}

				if deepEqual(resp.Request.Variables, reqBody.Variables) {
					jsonResp, err = json.Marshal(resp.Response)
					if err != nil {
						return responseGraphQLError(err, bodyBytes), nil
					}
				}
			}

			if len(jsonResp) == 0 {
				return responseGraphQLError(fmt.Errorf("query not found in prepared responses: %+v", preparedQueries), bodyBytes), nil
			}

			r := io.NopCloser(bytes.NewReader([]byte(jsonResp)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	})
}

// NewMockGraphQLClientSingle creates a mock graphql client with response map
// Find responses by query string without validating variables
func NewMockGraphQLClientQueries(responses map[string]string) *graphql.Client {
	preparedQueries := make([]string, 0, len(responses))
	for q := range responses {
		preparedQueries = append(preparedQueries, q)
	}
	return graphql.NewClient("/v1/graphql", &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return responseGraphQLError(err, []byte("io.ReadAll")), nil
			}

			var reqBody graphql.GraphQLRequestPayload
			err = json.Unmarshal(bodyBytes, &reqBody)

			if err != nil {
				return responseGraphQLError(err, bodyBytes), nil
			}

			var jsonResp string

			for q, resp := range responses {
				if strings.TrimSpace(q) == strings.TrimSpace(reqBody.OperationName) ||
					strings.TrimSpace(q) == strings.TrimSpace(reqBody.Query) {
					jsonResp = resp
					break
				}
			}
			if jsonResp == "" {
				return responseGraphQLError(fmt.Errorf("query not found in prepared responses: %+v", preparedQueries), bodyBytes), nil
			}

			r := io.NopCloser(bytes.NewReader([]byte(jsonResp)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	})
}

// NewMockGraphQLClientSingle creates a mock graphql client with static response
func NewMockGraphQLClientSingle(data any, err *graphql.Error) *graphql.Client {
	var jsonResp string
	if sResp, ok := data.(string); ok {
		jsonResp = sResp
	} else {
		jsonResp = EncodeMockGraphQLResponse(data, err)
	}
	return graphql.NewClient("/v1/graphql", &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			r := io.NopCloser(bytes.NewReader([]byte(jsonResp)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	})
}

// NewMockGraphQLAffectedRowsResponse returns json response of a mutation response with affectedRows
func NewMockGraphQLAffectedRowsResponse(mutationName string, affectedRows int) string {
	return EncodeMockGraphQLResponse(map[string]any{
		mutationName: map[string]int{
			"affected_rows": affectedRows,
		},
	}, nil)
}

type graphqlResponse struct {
	Data   any             `json:"data,omitempty"`
	Errors []graphql.Error `json:"errors,omitempty"`
}

// EncodeMockGraphQLResponse encodes data inputs to the graphql response payload
func EncodeMockGraphQLResponse(data any, err *graphql.Error) string {
	resp := graphqlResponse{}
	if data != nil {
		resp.Data = data
	}
	if err != nil {
		resp.Errors = []graphql.Error{*err}
	}

	bResp, encodeErr := json.Marshal(resp)
	if encodeErr != nil {
		panic(encodeErr)
	}

	return string(bResp)
}

func responseGraphQLError(err error, requestBody []byte) *http.Response {
	reqError := EncodeMockGraphQLResponse(nil, &graphql.Error{
		Message: err.Error(),
		Extensions: map[string]any{
			"code":         "validation-failed",
			"path":         "$.selectionSet.test",
			"request_body": string(requestBody),
		},
	})
	r := io.NopCloser(bytes.NewReader([]byte(reqError)))
	return &http.Response{
		StatusCode: 400,
		Body:       r,
	}
}

func deepEqual(v1, v2 interface{}) bool {
	if reflect.DeepEqual(v1, v2) {
		return true
	}
	var x1 interface{}
	bytesA, _ := json.Marshal(v1)
	_ = json.Unmarshal(bytesA, &x1)
	var x2 interface{}
	bytesB, _ := json.Marshal(v2)
	_ = json.Unmarshal(bytesB, &x2)
	if reflect.DeepEqual(x1, x2) {
		return true
	}
	return false
}
