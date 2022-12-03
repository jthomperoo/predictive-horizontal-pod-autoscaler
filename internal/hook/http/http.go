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

// Package http handles interactions over HTTP
package http

import (
	"context"
	"fmt"
	"io"
	gohttp "net/http"
	"strings"
	"time"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

// Type http represents an HTTP request
const Type = "http"

const (
	// BodyParameterMode is a configuration flag specifying the value passed via
	// HTTP should be through the HTTP body parameter
	BodyParameterMode = "body"
	// QueryParameterMode is a configuration flag specifying the value passed
	// via HTTP should be through the HTTP query parameter
	QueryParameterMode = "query"
)

const (
	// QueryParameterKey is the key of the query parameter passed if the query
	// parameter mode is used, in the form https://example.com?value="DATA"
	QueryParameterKey = "value"
)

// Execute represents a way to execute HTTP requests with values as parameters.
type Execute struct {
	Client gohttp.Client
}

// ExecuteWithValue executes an HTTP request with the value provided as
// parameter, configurable to be either in the body or query string
func (e *Execute) ExecuteWithValue(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
	if definition.HTTP == nil {
		return "", fmt.Errorf("missing required 'http' configuration on hook definition")
	}

	// Set up request using hook definition and URL provided
	req, err := gohttp.NewRequest(definition.HTTP.Method, definition.HTTP.URL, nil)
	if err != nil {
		return "", err
	}

	// Set parameter value, based on configuration option
	switch definition.HTTP.ParameterMode {
	case BodyParameterMode:
		// Set body parameter
		req.Body = io.NopCloser(strings.NewReader(value))
	case QueryParameterMode:
		// Set query parameter
		query := req.URL.Query()
		query.Add(QueryParameterKey, value)
		req.URL.RawQuery = query.Encode()
	default:
		return "", fmt.Errorf("unknown parameter mode '%s'", definition.HTTP.ParameterMode)
	}

	// Add headers
	for key, val := range definition.HTTP.Headers {
		req.Header.Add(key, val)
	}

	// Set up a context to provide an HTTP request timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(definition.Timeout)*time.Millisecond)
	defer cancel()

	// Make request
	resp, err := e.Client.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check for a successful response code
	success := false
	for _, successCode := range definition.HTTP.SuccessCodes {
		if resp.StatusCode == successCode {
			success = true
			break
		}
	}

	if !success {
		return "", fmt.Errorf("http request failed, status: [%d], response: '%s'", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// GetType returns the http executer type
func (e *Execute) GetType() string {
	return Type
}
