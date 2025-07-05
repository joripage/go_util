package shardqueue

import (
	"encoding/binary"
	"hash/fnv"
	"log"
	"math"
)

type Shardqueue struct {
	numShard  int
	queueSize int
	queue     []chan interface{}
}

type processFunc func(i interface{}) error

func NewShardQueue(numShard, queueSize int) *Shardqueue {
	sq := &Shardqueue{
		numShard:  numShard,
		queueSize: queueSize,
		queue:     make([]chan interface{}, numShard),
	}

	return sq
}

func (sq *Shardqueue) Start(fn processFunc) {
	for i := 0; i < sq.numShard; i++ {
		sq.queue[i] = make(chan interface{}, sq.queueSize)
		go sq.shardWorker(i, sq.queue[i], fn)
	}
}

func (sq *Shardqueue) Stop() {
	for i := 0; i < sq.numShard; i++ {
		close(sq.queue[i])
	}
}

func (sq *Shardqueue) Shard(routingKey interface{}, msg interface{}) {
	shard := hashKeyToShard(convertKeyToBytes(routingKey), sq.numShard)
	sq.queue[shard] <- msg
}

func (sq *Shardqueue) shardWorker(id int, ch chan interface{}, fn processFunc) {
	for msg := range ch {
		if err := fn(msg); err != nil {
			log.Printf("Shard %d process error: %v", id, err)
		}
	}
	log.Printf("Shard %d done", id)
}

func hashKeyToShard(key []byte, numShard int) int {
	h := fnv.New32a()
	h.Write(key)
	return int(h.Sum32()) % numShard
}

func convertKeyToBytes(key interface{}) []byte {
	switch v := key.(type) {
	case []byte:
		return v

	case string:
		return []byte(v)

	case int:
		return intToBytes(int64(v))
	case int32:
		return intToBytes(int64(v))
	case int64:
		return intToBytes(v)

	case uint:
		return uintToBytes(uint64(v))
	case uint32:
		return uintToBytes(uint64(v))
	case uint64:
		return uintToBytes(v)

	case float64:
		return floatToBytes(v)
	case float32:
		return floatToBytes(float64(v))

	default:
		return []byte("defaultRoutingKey")
	}
}

func intToBytes(n int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	return buf
}

func uintToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return buf
}

func floatToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, bits)
	return buf
}
