# Firetail Go Library

Want to use Firetail in your project? Good news! Soon, you'll be able to find packages containing middleware for various different frameworks in the [middlewares](./middlewares) directory, and examples of their use in [examples](./examples). :)



## Tests

[![codecov](https://codecov.io/gh/FireTail-io/firetail-go-lib/branch/main/graph/badge.svg?token=QZX8OSE964)](https://codecov.io/gh/FireTail-io/firetail-go-lib)

Automated testing is setup with the `testing` package, using [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) for shorthand assertions. You can run them with `go test`.



## XML Support

The Firetail Go library does not come with XML request & response body support out of the box. You will need to implement your own as an `openapi3filter.BodyDecoder` and pass it to Firetail as part of the `CustomBodyDecoders` field of the `Options` struct. See the following example for a minimal XML decoder setup using [sbabiv/xml2map](https://github.com/sbabiv/xml2map):

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

