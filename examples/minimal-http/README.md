# Minimal net/http Example

This is an example of a simple web server implemented using the `net/http` package, with a single `/health` endpoint as described by the included OpenAPIv3 [app-spec.yaml](./app-spec.yaml), using the Firetail middleware.



## Development Environment

[Install Golang](https://go.dev/doc/install), clone the repo, build the web server & run it:

```bash
git clone git@github.com:FireTail-io/firetail-go-lib.git
cd firetail-go-lib/examples/minimal-http
go build
./minimal-http
```

Curl it!

```bash
curl 'localhost:8080/health'
```

You should get the following response:

```
I'm Healthy!
```

And the logs of the server should (currently) read something a little like this:

```
2022/10/25 16:11:10 Sending 1 log(s) using API key ''...
2022/10/25 16:11:10 Entry #0: {"dateCreated":1666710669299,"executionTime":50,"request":{"body":"","headers":{"Accept":["*/*"],"User-Agent":["curl/7.79.1"]},"httpProtocol":"HTTP/1.1","ip":"127.0.0.1","method":"GET","uri":"http://localhost:8080/health","resource":"/health"},"response":{"body":"I'm Healthy! :)","headers":{"Content-Type":["text/plain"]},"statusCode":200},"version":"1.0.0-alpha"}
```

