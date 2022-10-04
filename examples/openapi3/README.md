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
curl localhost:8080
```

You should get the following response:

```
I'm Healthy!
```

And the logs of the server should (currently) read something a little like this:

```
2022/10/04 12:34:03 I'm Healthy!
2022/10/04 12:34:03 Request body to be sent to Firetail logging endpoint: {"version":"","dateCreated":1664883243702,"execution_time":0,"source_code":"","req":{"httpProtocol":"HTTP/1.1","url":"","headers":{"Accept":["*/*"],"User-Agent":["curl/7.79.1"]},"path":"/","method":"GET","oPath":"","fPath":"","args":{},"body":"","ip":"127.0.0.1:65050","pathParams":""},"resp":{"status_code":0,"content_len":0,"content_enc":"","failed_res_body":"","body":"","headers":{},"content_type":""}}
```

