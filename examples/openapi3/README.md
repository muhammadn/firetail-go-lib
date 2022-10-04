# Firetail OpenAPIv3 Prototype

This prototype currently has a single `/health` endpoint, implemented as described by the OpenAPIv3 [app-spec.yaml](./app-spec.yaml).



## Development Environment

[Install Golang](https://go.dev/doc/install), clone the repo, build the web server & run it:

```bash
git clone git@github.com:FireTail-io/firetail-go-lib.git
cd firetail-go-lib/examples/openapi3
go build -o openapi3-example
./openapi3-example
```

Curl it!

```bash
curl 'localhost:8080/example-path?example-param=example-value' -H 'Content-Type: application/json' -d '{"example":"body"}'
```

You should get the following response:

```
I'm Healthy!
```

And the logs of the server should (currently) read something a little like this:

```
2022/10/04 16:11:53 I'm Healthy, and I take some time!
2022/10/04 16:11:53 Request body to be sent to Firetail logging endpoint: {
        "version": "",
        "dateCreated": 1664896313861,
        "executionTime": 50,
        "sourceCode": "",
        "request": {
                "httpProtocol": "HTTP/1.1",
                "url": "localhost:8080/example-path?example-param=example-value",
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
                "method": "POST",
                "body": "{\"example\":\"body\"}",
                "ip": "127.0.0.1:50562"
        },
        "response": {
                "statusCode": 200,
                "contentLength": 34,
                "body": "I'm Healthy, and I take some time!",
                "headers": {
                        "Content-Type": [
                                "text/html"
                        ]
                }
        }
}
```

