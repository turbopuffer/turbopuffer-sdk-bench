# turbopuffer SDK Benchmarks

This repository contains benchmarks for measuring the performance
of different turbopuffer SDKs.

Available benchmarks:

  * [`python`](./python)

## Details

* Benchmarks generally measure both upsert and query performance.
* Benchmarks are single-threaded.
* Benchmarks include the latency of talking to the turbopuffer API, so
  ideally run them from the same cloud provider and region as the turbopuffer
  cluster you're accessing.
