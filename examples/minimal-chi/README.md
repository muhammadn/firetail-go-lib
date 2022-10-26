# Minimal Chi Example

This example Chi app is a copy of the [hello world example](https://github.com/go-chi/chi/tree/master/_examples/hello-world) from the [go-chi/chi repo](https://github.com/go-chi/chi), which has been modified to use Firetail middleware and conform to the included [app-spec.yaml](./app-spec.yaml).



## Development Environment

[Install Golang](https://go.dev/doc/install), clone the repo, build the web server & run it:

```bash
git clone git@github.com:FireTail-io/firetail-go-lib.git
cd firetail-go-lib/examples/minimal-chi
go build
./minimal-chi
```

Curl it!

```bash
curl localhost:3333/health
```

You should get the following response:

```
I'm Healthy! :)
```

And the logs should (currently) look like this:

```
2022/10/25 16:07:25 [Joshuas-MBP/Bzeaxai9bb-000001] "GET http://localhost:3333/health HTTP/1.1" from 127.0.0.1:50970 - 200 15B in 13.041µs
2022/10/25 16:07:26 Sending 1 log(s) using API key ''...
2022/10/25 16:07:26 Entry #0: {"dateCreated":1666710445229,"executionTime":0,"request":{"body":"","headers":{"Accept":["*/*"],"User-Agent":["curl/7.79.1"]},"httpProtocol":"HTTP/1.1","ip":"127.0.0.1","method":"GET","uri":"http://localhost:3333/health","resource":"/health"},"response":{"body":"I'm Healthy! :)","headers":{"Content-Type":["text/plain"]},"statusCode":200},"version":"1.0.0-alpha"}
```

