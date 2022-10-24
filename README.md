# Firetail Go Library

Want to use Firetail in your project? Good news! Soon, you'll be able to find packages containing middleware for various different frameworks in the [middlewares](./middlewares) directory, and examples of their use in [examples](./examples). :)



## Tests

[![codecov](https://codecov.io/gh/FireTail-io/firetail-go-lib/branch/main/graph/badge.svg?token=QZX8OSE964)](https://codecov.io/gh/FireTail-io/firetail-go-lib)

Automated testing is setup with the `testing` package, using [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) for shorthand assertions. You can run them with `go test`.



## XML Support

The Firetail Go library does not come with XML request & response body support out of the box. You will need to implement your own as an `openapi3filter.BodyDecoder` and pass it to Firetail as part of the `CustomBodyDecoders` field of the `firetail.Options` struct. See the following example for a minimal XML decoder setup using [sbabiv/xml2map](https://github.com/sbabiv/xml2map):

```go
middleware, err := GetMiddleware(&Options{
	OpenapiSpecPath: "./app-spec.yaml",
	CustomBodyDecoders: map[string]openapi3filter.BodyDecoder{
		"application/xml": func(r io.Reader, h http.Header, sr *openapi3.SchemaRef, ef openapi3filter.EncodingFn) (interface{}, error) {
			return xml2map.NewDecoder(r).Decode()
		},
	},
})
```



## Authentication

If you use `securitySchemes` in your OpenAPI specification, you will need to populate the `firetail.Options` struct's `AuthCallback` field with your authentication logic. It is your responsibility to route the names of the security schemes in your OpenAPI specification to the appropriate authentication logic.

For example, for the following `securitySchemes`:

```yaml
components:
  securitySchemes:
    BasicAuth:
      type: http
      scheme: basic
```

Your `AuthCallback` could look like this:

```go
AuthCallback: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
  switch ai.SecuritySchemeName {
    case "MyBasicAuth":
    return handleBasicAuth(ctx, ai)
    default:
    return errors.New("security scheme not implemented")
  }
},
```



### Custom Auth Error Responses

In order to customise the errors returned by your application when a request fails to authenticate, you can pick up the errors returned by your `AuthCallback` in a custom `ErrHandler`. This allows you to, for example, add the `WWW-Authenticate` header on responses to requests that fail to validate against a basic auth security requirement:

```go
// We'll use this err when the basic auth fails to validate.
var BasicAuthErr = errors.New("invalid authorization token")

firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
	OpenapiSpecPath: "app-spec.yaml",

	// First, let's write our auth callback which, if the name of the security scheme it's being asked to check
	// is 'MyBasicAuth', will check that the Authorization header contains the b64 encoding of 'admin:password'.
	AuthCallback: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
		switch ai.SecuritySchemeName {
		case "MyBasicAuth":
			token := ai.RequestValidationInput.Request.Header.Get("Authorization")
			if token != "Basic YWRtaW46cGFzc3dvcmQ=" {
				return BasicAuthErr
			}
			return nil
		default:
			return errors.New("security scheme not implemented")
		}
	},

	// Then, in our ErrCallback, we can check if the error it's received is a security error. If it is, and 
	// one of its suberrs is a BasicAuthErr, then we can add the WWW-Authenticate header to the response.
	ErrCallback: func(err firetail.ErrorAtRequest, w http.ResponseWriter, r *http.Request) {
		if securityErr, isSecurityErr := err.(firetail.ErrorAuthNoMatchingSchema); isSecurityErr {
			for _, subErr := range securityErr.Err.Errors {
				if subErr == BasicAuthErr {
					w.Header().Add("WWW-Authenticate", "Basic")
					break
				}
			}
		}
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(err.StatusCode())
		w.Write([]byte(err.Error()))
	},	
})
```