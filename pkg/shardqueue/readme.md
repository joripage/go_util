# Shard queue

A small package to shard message into queue

## Installation

```bash
go get github.com/joripage/go_util/pkg/shardqueue
```

## How to use

Code in cmd/shardqueue

```go
 sq := shardqueue.NewShardQueue(numShard, queueSize)
 sq.Start(func(msg interface{}) error {
  log.Println("process msg", msg)
  return nil
 })
```

## Why using fnv

- One of FNV's key advantages is that it is very simple to implement. Start with an initial hash value of FNV offset basis. For each byte in the input, multiply hash by the FNV prime, then XOR it with the byte from the input. The alternate algorithm, FNV-1a, reverses the multiply and XOR steps. (link: <https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function>)
- fnv.New32a() is one of the fastest function in Go stdlib

```go
BenchmarkFnv32a-8     195 ns/op
BenchmarkCRC32-8      230 ns/op
BenchmarkSHA1-8       1400 ns/op
BenchmarkSHA256-8     2000 ns/op
BenchmarkMD5-8        1100 ns/op
```

- We can use xxhash if you need faster, or use sha256 if you need safe

| Hash           | speed         | Distribution     | Usage        | Safe    |
| -------------- | ------------- | ---------------- | -------------| ------- |
| `fnv.New32a()` |  Fast         |  Good            |  Easy        |  NO     |
| `xxhash`       |  Very fast    |  Very good       | ️ Need lib    |  NO     |
| `sha256`       |  Slow         |  Evenly          |  Complicated |  YES    |
