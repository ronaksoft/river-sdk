# Rony (Fast and Scalable RPC Framework)

![](https://img.shields.io/github/stars/ronaksoft/rony?color=808080)
![](https://img.shields.io/github/license/ronaksoft/rony?color=808080)

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/ronaksoft/rony?include_prereleases&logoColor=808080)
![](https://github.com/ronaksoft/rony/actions/workflows/go.yml/badge.svg?color=379c9c&style=flat-square)
![](https://github.com/ronaksoft/rony/actions/workflows/devskim-analysis.yml/badge.svg?color=379c9c&style=flat-square)
![](https://github.com/ronaksoft/rony/actions/workflows/codacy-analysis.yml/badge.svg?color=379c9c&style=flat-square)

## About
**Rony** lets you create a clustered aware service easily. Checkout [Wiki](https://rony.ronaksoft.com) 



### Performance

Rony is very fast and with very low overhead. In a non-raft environment it adds < 25us latency, and in raft-enabled scenario around 1ms latency.

Rony has negligible memory footprint by reusing buffers and pooling.

```commandline
BenchmarkEdgeServerMessageSerial
BenchmarkEdgeServerMessageSerial-16                       901370              1195 ns/op              87 B/op          2 allocs/op
BenchmarkEdgeServerMessageParallel
BenchmarkEdgeServerMessageParallel-16                    4521645               272 ns/op              83 B/op          2 allocs/op
BenchmarkEdgeServerWithRaftMessageSerial
BenchmarkEdgeServerWithRaftMessageSerial-16                 9541            132065 ns/op            5034 B/op        116 allocs/op
BenchmarkEdgeServerWithRaftMessageParallel
BenchmarkEdgeServerWithRaftMessageParallel-16             124658              8438 ns/op            4462 B/op         51 allocs/op

BenchmarkServerWithWebsocket-16            46514             25138 ns/op             691 B/op         19 allocs/op
```

---

### Easy Setup for advanced scenarios

```go
package main

import (
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/edge"
	"os"
)

func main() {
	serverID := "UniqueServerID"
	edgeServer := edge.NewServer(serverID,
		edge.WithTcpGateway(edge.TcpGatewayConfig{
			Concurrency:   1000,
			MaxIdleTime:   0,
			ListenAddress: "0.0.0.0:80",
			ExternalAddrs: []string{"127.0.0.1:80"}, // ExternalAddr could be used when the server is behind proxy or nats
		}),
		edge.WithUdpTunnel(edge.UdpTunnelConfig{
			ListenAddress: "0.0.0.0:8080",
			MaxBodySize:   0,
			ExternalAddrs: nil,
		}),
		edge.WithGossipCluster(edge.GossipClusterConfig{
			Bootstrap:  true,
			ReplicaSet: 1,
			GossipPort: 7091,
		}),
	)

	service.RegisterSampleService(&service.SampleService{}, edgeServer)
	edgeServer.Start()

	edgeServer.ShutdownWithSignal(os.Kill)
}
```

**This code does not run, please check example directory for working examples**


---

## Shoulders

Rony is made of big and popular packages. Without these great libraries building Rony was not possible.

* [Hashicorp MemberList](https://github.com/hashicorp/memberlist)
* [SPF13 Cobra](https://github.com/spf13/cobra)
* [SPF13 Viper](https://github.com/spf13/viper)
* [Valyala FastHttp](https://github.com/valyala/fasthttp)
* [Uber Zap](https://go.uber.org/zap)
* [Gobwas Websocket](https://github.com/gobwas/ws)
* [Allegro BigCache](https://github.com/allegro/bigcache)
* [GoBuffalo Genny](https://github.com/gobuffalo/genny)
* [GoBuffalo Plush](https://github.com/gobuffalo/plush)

### Contribution

We need to make a clear and understandable documentation for Rony, so any help would be appreciated. We also appreciate benchmarking Rony against other platforms for common
scenarios and will be cited.

### TODOs

- [x] Auto-generate REST handlers
- [ ] Auto-generate test units
- [ ] Implement scenario tests using `edgetest` framework
- [ ] Write Examples and Tutorials
