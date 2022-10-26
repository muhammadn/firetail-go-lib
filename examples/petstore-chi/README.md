# Petstore Chi Example

This project is a copy of the [oapi-codegen](https://github.com/deepmap/oapi-codegen) module's [petstore chi](https://github.com/deepmap/oapi-codegen/tree/master/examples/petstore-expanded/chi) example, which has had the oapi-codegen's request validation middleware replaced with the Firetail middleware.

Upon installing the Firetail middleware, several discrepancies were immediately discovered between the implementation and its corresponding [openapi 3.0 appspec](https://github.com/deepmap/oapi-codegen/blob/master/examples/petstore-expanded/petstore-expanded.yaml):

- The `GET /pets`, `POST /pets`, `GET /pets/{id}` and all error responses were missing their `Content-Type` headers. This is not mandatory, but it is good practice. This was rectified by adding the `Content-Type` headers to the responses.
- The `GET /pets` endpoint returned `null` when there were no pets in the store, but the openapi schema did not describe the response body as nullable. This was rectified by ammending the appspec to describe the response body as nullable.
- The `POST /pets` response code upon successful creation of a new pet was `201`, however the appspec defined a `200` response. This was rectified by ammending the appspec to define a `201` response, instead of a `200`, as `201` is a more apropriate status code in this case.



## Feature Overview

Using this familiar petstore example, we can demonstrate a number of the Firetail middleware's features:

1. [Request Validation](#request-validation)
2. [Authentication](#authentication)



### Request Validation

The Firetail middleware blocks:

- Requests made to paths that aren't defined in your appspec:
  ```bash
  curl localhost:8080/owners -v
  ```

  ```
  < HTTP/1.1 404 Not Found
  < Content-Type: text/plain
  [firetail] no matching path found for "/owners"
  ```

- Requests made to paths that are defined in your appspec, but are made with unsupported methods:
  ```bash
  curl localhost:8080/pets -X DELETE -v
  ```

  ```
  < HTTP/1.1 405 Method Not Allowed
  < Content-Type: text/plain
  [firetail] "/pets" path does not support DELETE method
  ```

- Requests made to paths defined in your appspec, with methods defined in your appspec, but with a `Content-Type` that hasn't been defined in your appspec:
  ```bash
  curl localhost:8080/pets -X POST -H "Content-Type: application/xml" -d '<?xml version="1.0" encoding="UTF-8" ?><root><name>Hector</name></root>' -v
  ```

  ```
  < HTTP/1.1 415 Unsupported Media Type
  firetail - /pets route does not support content type application/xml
  ```

- Requests made with path parameters that don't match the schema defined in your appspec:
  ```bash
  curl localhost:8080/pets/abc -v
  ```

  ```
  < HTTP/1.1 400 Bad Request
  < Content-Type: text/plain
  [firetail] request path parameter invalid: parameter "id" in path has an error: value abc: an invalid integer: invalid syntax
  ```

- Requests made with query parameters that don't match the schema defined in your appspec:
  ```bash
  curl localhost:8080/pets?limit=abc -v
  ```

  ```
  < HTTP/1.1 400 Bad Request
  < Content-Type: text/plain
  [firetail] request query parameter invalid: parameter "limit" in query has an error: value abc: an invalid integer: invalid syntax
  ```

- Requests made with bodies that don't match the schema defined in your appspec:

  ```bash
  curl localhost:8080/pets -X POST -H "Content-Type: application/json" -d '{"name":123}' -v
  ```

  ```
  < HTTP/1.1 400 Bad Request
  < Content-Type: text/plain
  [firetail] request body invalid: request body has an error: doesn't match the schema: Error at "/name": Field must be set to string or not be present
  Schema:
    {
      "description": "Name of the pet",
      "type": "string"
    }
  
  Value:
    "number, integer"
  ```



### Authentication

The petstore appspec has been modified to apply a security scheme to the `DELETE /pets/{id}` endpoint. This was implemented by simply defining an `AuthCallback` and providing it to the Firetail middleware as follows:

```go
	firetailMiddleware, err := firetail.GetMiddleware(&firetail.Options{
		OpenapiSpecPath: "./petstore-expanded.yaml",
		AuthCallback: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
			switch ai.SecuritySchemeName {
			case "MyBearerAuth":
				tokenParts := strings.Split(ai.RequestValidationInput.Request.Header.Get("Authorization"), " ")
				if len(tokenParts) < 2 || tokenParts[1] != "header.payload.signature" {
					return errors.New("invalid bearer token for MyBearerAuth")
				}
				return nil
			default:
				return errors.New(fmt.Sprintf("security scheme \"%s\" not implemented", ai.SecuritySchemeName))
			}
		},
	})
```

We can test this out by creating a pet:

```bash
curl localhost:8080/pets -X POST -H "Content-Type: application/json" -d '{"name":"Hector"}' -v
```

```
< HTTP/1.1 201 Created
< Content-Type: application/json
{"id":1001,"name":"Hector"}
```

Then attempting to:

1. Delete the pet without a JWT:
   ```bash
   curl localhost:8080/pets/1000 -X DELETE -v
   ```

   ```
   < HTTP/1.1 401 Unauthorized
   < Content-Type: text/plain
   firetail - request did not satisfy security requirements: Security requirements failed, errors: no bearer token supplied for "MyBearerAuth"
   ```

2. Delete the pet with an invalid JWT:
   ```bash
   curl localhost:8080/pets/1000 -X DELETE -H "Authorization: bearer header.payload.badsignature" -v
   ```

   ```
   < HTTP/1.1 401 Unauthorized
   < Content-Type: text/plain
   firetail - request did not satisfy security requirements: Security requirements failed, errors: invalid bearer token for "MyBearerAuth"
   ```

3. Delete the pet with a valid JWT:
   ```bash
   curl localhost:8080/pets/1000 -X DELETE -H "Authorization: bearer header.payload.signature" -v
   ```

   ```
   < HTTP/1.1 204 No Content
   ```

   After which there should be no pets:

   ```bash
   curl localhost:8080/pets -v
   ```

   ```
   < HTTP/1.1 200 OK
   < Content-Type: application/json
   null
   ```