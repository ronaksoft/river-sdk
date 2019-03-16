Changelog from v3.0.1 and up. Prior changes don't have a changelog.

# v3.2.0

* Add `StreamReader` type to make working with redis' new [Stream][stream]
  functionality easier.

* Make `Sentinel` properly respond to `Client` method calls. Previously it
  always created a new `Client` instance when a secondary was requested, now it
  keeps track of instances internally.

* Make default `Dial` call have a timeout for connect/read/write. At the same
  time, normalize default timeout values across the project.

* Implicitly pipeline commands in the default Pool implementation whenever
  possible. This gives a throughput increase of nearly 5x for a normal parallel
  workload.

[stream]: https://redis.io/topics/streams-intro

# v3.1.0

* Add support for marshaling/unmarshaling structs.

# v3.0.1

* Make `Stub` support `Pipeline` properly.
