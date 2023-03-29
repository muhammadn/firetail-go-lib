# Firetail Go Library

[![License](https://img.shields.io/pypi/l/firetail.svg)](https://github.com/FireTail-io/firetail-go-lib/blob/main/LICENSE.txt) [![Go Reference](https://pkg.go.dev/badge/github.com/FireTail-io/firetail-go-lib#section-readme.svg)](https://pkg.go.dev/github.com/FireTail-io/firetail-go-lib#section-readme) [![codecov](https://codecov.io/gh/FireTail-io/firetail-go-lib/branch/main/graph/badge.svg?token=QZX8OSE964)](https://codecov.io/gh/FireTail-io/firetail-go-lib)

Middlewares providing request and response validation against an OpenAPI spec at runtime, optionally integrating with the Firetail SaaS. Packages containing middleware for various different frameworks can be found in the [middlewares](./middlewares) directory, and examples of their use in [examples](./examples).



## Getting Started

### Middleware for `net/http`

Get the middleware:

```bash
go get github.com/FireTail-io/firetail-go-lib/middlewares/http
```

Import it:

```go
import firetail "github.com/FireTail-io/firetail-go-lib/middlewares/http"
```

Create a middleware using `GetMiddleware`; see the `Options` struct for all the available configurations:

```go
firetailMiddleware, err := firetail.GetMiddleware(
	&firetail.Options{
		OpenapiSpecPath: path,
		LogApiKey:       apiToken,
	},
)
if err != nil {
	// Handle the err...
}
```

You will then have a `func(next http.Handler) http.Handler`, `firetailMiddleware`, which you can use to wrap a `http.Handler` just the same as with the middleware from [`net/http/middleware`](https://pkg.go.dev/go.ntrrg.dev/ntgo/net/http/middleware). This should also be suitable for [Chi](https://go-chi.io/#/pages/middleware).



## Tests

Automated testing is setup with the `testing` package, using [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) for shorthand assertions. You can run them with `go test`.



## XML Support

The Firetail Go library does not come with XML request & response body decoding support out of the box. You will need to implement your own decoder as an [openapi3filter.BodyDecoder](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3filter#BodyDecoder) and pass it to Firetail as part of the `CustomBodyDecoders` field of the `firetail.Options` struct. See the following example for a minimal XML decoder setup using [sbabiv/xml2map](https://github.com/sbabiv/xml2map):

```go
middleware, err := firetail.GetMiddleware(&firetail.Options{
	OpenapiSpecPath: "./app-spec.yaml",
	CustomBodyDecoders: map[string]openapi3filter.BodyDecoder{
		"application/xml": func(r io.Reader, h http.Header, sr *openapi3.SchemaRef, ef openapi3filter.EncodingFn) (interface{}, error) {
			return xml2map.NewDecoder(r).Decode()
		},
	},
})
```



## Authentication

If you use `securitySchemes` in your OpenAPI specification, you will need to populate the `firetail.Options` struct's `AuthCallbacks` field with a callback for each security scheme implementing your authentication logic.

For example, for the following `securitySchemes`:

```yaml
components:
  securitySchemes:
    MyBasicAuth:
      type: http
      scheme: basic
```

Your `AuthCallback` could look like this:

```go
AuthCallbacks: map[string]func(context.Context, *openapi3filter.AuthenticationInput){
	"MyBasicAuth": func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
		token := ai.RequestValidationInput.Request.Header.Get("Authorization")
		return validateBasicAuthToken(token)
	},
},
```



### Custom Auth Error Responses

In order to customise the errors returned by your application when a request fails to authenticate, you can pick up the errors returned by your `AuthCallbacks` in a custom `ErrHandler`. This allows you to, for example, add the `WWW-Authenticate` header on responses to requests that fail to validate against a basic auth security requirement:

```go
// We'll use this err when the basic auth fails to validate.
var BasicAuthErr = errors.New("invalid authorization token")

firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
	OpenapiSpecPath: "app-spec.yaml",

	// First, let's write our auth callback which, if the name of the security scheme it's being asked to check
	// is 'MyBasicAuth', will check that the Authorization header contains the b64 encoding of 'admin:password'.
	AuthCallbacks: map[string]func(context.Context, *openapi3filter.AuthenticationInput){
		"MyBasicAuth": func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
			token := ai.RequestValidationInput.Request.Header.Get("Authorization")
			if token != "Basic YWRtaW46cGFzc3dvcmQ=" {
				return BasicAuthErr
			}
			return nil
		},
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