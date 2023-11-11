# raknet-proxy

`raknet-proxy` proxies an upstream RakNet server. To the downstream client, it looks as though the proxy server is the RakNet server it is communicating with, and to the upstream RakNet server, the proxy appears to be the client.

This tool was designed to allow proxying of Rust from a GeForce NOW instance via a server with better routing properties. GeForce NOW instances are unprivileged, so it's not possible to do something like run a VPN, therefore all proxying must happen "in game". Using `raknet-proxy` allows to run `connect <proxy ip>` from within the Rust console and have the game connect to the upstream game server via the proxy.

```
go run ./cmd/raknet-proxy --listen-port 28016 --log-format text --log-level trace --server-hostname localhost --server-port 28017
```
```
go run ./cmd/mirror --log-format text --log-level trace --listen-port 28017
```
```
printf "hello world" | socat - UDP-DATAGRAM:172.17.0.2:28016
```
