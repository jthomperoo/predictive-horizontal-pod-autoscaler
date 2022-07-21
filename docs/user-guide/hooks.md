# Hooks

Hooks specify how user logic should be called by the Predictive Horizontal Pod Autoscaler.

## http

The http hook allows defining an HTTP request for the autoscaler to make. Any relevant information will be provided to
the target of the request by HTTP parameters - either `query` or `body` parameters. An error is signified by a status
code that is not defined to be successful in the configuration; if this kind of error occurs the autoscaler will
capture the response body and log it.

### Example

This is an example configuration of the http hook for runtime tuning fetching with Holt Winters:

```yaml
holtWinters:
  runtimeTuningFetchHook:
    type: "http"
    timeout: 2500
    http:
      method: "GET"
      url: "https://www.jamiethompson.me"
      successCodes:
        - 200
      headers:
        exampleHeader: exampleHeaderValue
      parameterMode: query
```

Breaking this example down:

- `type` = the type of the hook, for this example it is an `http` hook.
- `timeout` = the maximum time the hook can take in milliseconds, for this
  example it is `2500` (2.5 seconds), if it takes longer than this it will count
  the hook as failing.
- `http` = configuration of the HTTP request.
  - `method` = the HTTP method of the HTTP request.
  - `url` = the URL to target with the HTTP request.
  - `successCodes` = a list of success codes defining how to determine if the
    request is successful - if the request responds with a code not on this list
    it will be assumed to be a failure.
  - `headers` = a dictionary of headers that can be provided with the request,
    in this example the key is `exampleHeader` and the value is
    `exampleHeaderValue`. This is an optional parameter.
  - `parameterMode` = the mode for passing parameters to the target; either
    `query` - as a query parameter, or `body` - as a body parameter. In this
    example it is by query parameter.

### POST Example

This is an example using HTTP `POST` and information passed as a body parameter.

```yaml
runtimeTuningFetchHook:
  type: "http"
  timeout: 2500
  http:
    method: "POST"
    url: "https://www.jamiethompson.me"
    successCodes:
      - 200
      - 202
    parameterMode: body
```
