# Hub

The HTTP server that channels the captured traffic and traces using WebSockets to the client. The central
piece of software that [workers](https://github.com/kubeshark/worker) communicate with.

## Go build

Build:

```shell
go build -o hub .
```

Run:

```shell
./hub -port 8898
```
