# Firetail Chi Example

This example Chi app is a copy of the [hello world example](https://github.com/go-chi/chi/tree/master/_examples/hello-world) from the [go-chi/chi repo](https://github.com/go-chi/chi), which has been modified to use Firetail middleware and conform to the included [app-spec.yaml](./app-spec.yaml). You can follow the git history to see the changes made.



## Development Environment

[Install Golang](https://go.dev/doc/install), clone the repo, build the web server & run it:

```bash
git clone git@github.com:FireTail-io/firetail-go-lib.git
cd firetail-go-lib/examples/chi
go build
./chi
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
2022/10/14 14:40:56 [Joshuas-MacBook-Pro.local/ok1lZH3C0a-000001] "GET http://localhost:3333/health HTTP/1.1" from 127.0.0.1:55881 - 200 15B in 4.167µs
2022/10/14 14:40:57 Sending 1 log(s) to ''...
2022/10/14 14:40:57 Entry #0: {"dateCreated":1665751256698,"executionTime":0,"request":{"body":"","headers":{"Accept":["*/*"],"User-Agent":["curl/7.79.1"]},"httpProtocol":"HTTP/1.1","ip":"127.0.0.1","method":"GET","uri":"http://localhost:3333/health","resource":"/health"},"response":{"body":"I'm Healthy! :)","headers":{"Content-Type":["text/plain"]},"statusCode":200},"version":"1.0.0-alpha"}
```

