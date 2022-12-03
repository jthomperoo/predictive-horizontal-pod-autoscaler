/*
Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http_test

import (
	"errors"
	"fmt"
	"io"
	gohttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook/http"
)

type testHTTPClient struct {
	RoundTripReactor func(req *gohttp.Request) (*gohttp.Response, error)
}

func (f *testHTTPClient) RoundTrip(req *gohttp.Request) (*gohttp.Response, error) {
	return f.RoundTripReactor(req)
}

type testReader struct {
	ReadReactor  func(p []byte) (n int, err error)
	CloseReactor func() error
}

func (f *testReader) Read(p []byte) (n int, err error) {
	return f.ReadReactor(p)
}

func (f *testReader) Close() error {
	return f.CloseReactor()
}

func TestExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    string
		expectedErr error
		definition  *jamiethompsonmev1alpha1.HookDefinition
		value       string
		execute     http.Execute
	}{
		{
			description: "Fail, missing HTTP method configuration",
			expected:    "",
			expectedErr: errors.New(`missing required 'http' configuration on hook definition`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
			},
			value:   "test",
			execute: http.Execute{},
		},
		{
			description: "Fail, invalid HTTP method",
			expected:    "",
			expectedErr: errors.New(`net/http: invalid method "*?"`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method: "*?",
					URL:    "https://custompodautoscaler.com",
				},
			},
			value:   "test",
			execute: http.Execute{},
		},
		{
			description: "Fail, unknown parameter mode",
			expected:    "",
			expectedErr: errors.New(`unknown parameter mode 'unknown'`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "unknown",
				},
			},
			value:   "test",
			execute: http.Execute{},
		},
		{
			description: "Fail, request fail",
			expected:    "",
			expectedErr: errors.New(`Get "https://custompodautoscaler.com?value=test": Test network error!`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {
							return nil, errors.New("Test network error!")
						},
					},
				},
			},
		},
		{
			description: "Fail, timeout",
			expected:    "",
			expectedErr: errors.New(`Get "https://custompodautoscaler.com?value=test": context deadline exceeded`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
				Timeout: 5,
			},
			value: "test",
			execute: http.Execute{
				Client: func() gohttp.Client {
					testserver := httptest.NewServer(gohttp.HandlerFunc(func(rw gohttp.ResponseWriter, req *gohttp.Request) {
						time.Sleep(10 * time.Millisecond)
						// Send response to be tested
						rw.Write([]byte(`OK`))
					}))
					defer testserver.Close()

					return *testserver.Client()
				}(),
			},
		},
		{
			description: "Fail, invalid response body",
			expected:    "",
			expectedErr: errors.New(`Fail to read body!`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {
							resp := &gohttp.Response{
								Body: &testReader{
									ReadReactor: func(p []byte) (n int, err error) {
										return 0, errors.New("Fail to read body!")
									},
								},
								Header: gohttp.Header{},
							}
							resp.Header.Set("Content-Length", "1")
							return resp, nil
						},
					},
				},
			},
		},
		{
			description: "Fail, bad response code",
			expected:    "",
			expectedErr: errors.New(`http request failed, status: [400], response: 'bad request!'`),
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {
							return &gohttp.Response{
								Body:       io.NopCloser(strings.NewReader("bad request!")),
								Header:     gohttp.Header{},
								StatusCode: 400,
							}, nil
						},
					},
				},
			},
		},
		{
			description: "Success, POST, body parameter, 3 headers",
			expected:    "Success!",
			expectedErr: nil,
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "POST",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "body",
					Headers: map[string]string{
						"a": "testa",
						"b": "testb",
						"c": "testc",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {

							if !cmp.Equal(req.Method, "POST") {
								return nil, fmt.Errorf("Invalid method, expected 'POST', got '%s'", req.Method)
							}

							// Read the request body
							body, err := io.ReadAll(req.Body)
							if err != nil {
								return nil, err
							}

							if !cmp.Equal(req.Header.Get("a"), "testa") {
								return nil, fmt.Errorf("Missing header 'a'")
							}
							if !cmp.Equal(req.Header.Get("b"), "testb") {
								return nil, fmt.Errorf("Missing header 'a'")
							}
							if !cmp.Equal(req.Header.Get("c"), "testc") {
								return nil, fmt.Errorf("Missing header 'a'")
							}
							if !cmp.Equal(string(body), "test") {
								return nil, fmt.Errorf("Invalid body, expected 'test', got '%s'", body)
							}

							return &gohttp.Response{
								Body:       io.NopCloser(strings.NewReader("Success!")),
								Header:     gohttp.Header{},
								StatusCode: 200,
							}, nil
						},
					},
				},
			},
		},
		{
			description: "Success, GET, query parameter, 1 header",
			expected:    "Success!",
			expectedErr: nil,
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					Headers: map[string]string{
						"a": "testa",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {

							if !cmp.Equal(req.Method, "GET") {
								return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
							}

							query := req.URL.Query()

							if !cmp.Equal(query.Get("value"), "test") {
								return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
							}

							if !cmp.Equal(req.Header.Get("a"), "testa") {
								return nil, fmt.Errorf("Missing header 'a'")
							}

							return &gohttp.Response{
								Body:       io.NopCloser(strings.NewReader("Success!")),
								Header:     gohttp.Header{},
								StatusCode: 200,
							}, nil
						},
					},
				},
			},
		},
		{
			description: "Success, GET, query parameter, 0 headers",
			expected:    "Success!",
			expectedErr: nil,
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {

							if !cmp.Equal(req.Method, "GET") {
								return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
							}

							query := req.URL.Query()

							if !cmp.Equal(query.Get("value"), "test") {
								return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
							}

							return &gohttp.Response{
								Body:       io.NopCloser(strings.NewReader("Success!")),
								Header:     gohttp.Header{},
								StatusCode: 200,
							}, nil
						},
					},
				},
			},
		},
		{
			description: "Success, PUT, body parameter, 0 headers",
			expected:    "Success!",
			expectedErr: nil,
			definition: &jamiethompsonmev1alpha1.HookDefinition{
				Type: "http",
				HTTP: &jamiethompsonmev1alpha1.HTTPHook{
					Method:        "PUT",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "body",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			value: "test",
			execute: http.Execute{
				Client: gohttp.Client{
					Transport: &testHTTPClient{
						func(req *gohttp.Request) (*gohttp.Response, error) {

							if !cmp.Equal(req.Method, "PUT") {
								return nil, fmt.Errorf("Invalid method, expected 'PUT', got '%s'", req.Method)
							}

							// Read the request body
							body, err := io.ReadAll(req.Body)
							if err != nil {
								return nil, err
							}

							if !cmp.Equal(string(body), "test") {
								return nil, fmt.Errorf("Invalid body, expected 'test', got '%s'", body)
							}

							return &gohttp.Response{
								Body:       io.NopCloser(strings.NewReader("Success!")),
								Header:     gohttp.Header{},
								StatusCode: 200,
							}, nil
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.execute.ExecuteWithValue(test.definition, test.value)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestExecute_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		execute     http.Execute
	}{
		{
			description: "Return type",
			expected:    "http",
			execute:     http.Execute{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := test.execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
