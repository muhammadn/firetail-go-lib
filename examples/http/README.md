# Firetail OpenAPIv3 Prototype

This is an example of a simple web server implemented using the `net/http` package, with a single `/health` endpoint as described by the included OpenAPIv3 [app-spec.yaml](./app-spec.yaml).



## Development Environment

[Install Golang](https://go.dev/doc/install), clone the repo, build the web server & run it:

```bash
git clone git@github.com:FireTail-io/firetail-go-lib.git
cd firetail-go-lib/examples/http
go build -o http-example
./http-example
```

Curl it!

```bash
curl 'localhost:8080/health?example-param=example-value' -H 'Content-Type: application/json' -d '{"example":"body"}' -X GET
```

You should get the following response:

```
I'm Healthy!
```

And the logs of the server should (currently) read something a little like this:

```
2022/10/05 11:30:05 Request body to be sent to Firetail logging endpoint: {
        "dateCreated": 1664965805828,
        "executionTime": 51,
        "request": {
                "body": "{\"example\":\"body\"}",
                "headers": {
                        "Accept": [
                                "*/*"
                        ],
                        "Content-Length": [
                                "18"
                        ],
                        "Content-Type": [
                                "application/json"
                        ],
                        "User-Agent": [
                                "curl/7.79.1"
                        ]
                },
                "httpProtocol": "HTTP/1.1",
                "ip": "127.0.0.1",
                "method": "GET",
                "uri": "http://localhost:8080/health?example-param=example-value"
        },
        "response": {
                "body": "I'm Healthy!",
                "headers": {
                        "Content-Type": [
                                "text/plain"
                        ]
                },
                "statusCode": 200
        },
        "version": "1.0.0-alpha"
}
```

