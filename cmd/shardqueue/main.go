package main

import (
	"go_util/pkg/shardqueue"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	numShard := 8
	queueSize := 1000
	totalMsg := int32(10000)
	processCount := int32(0)
	begin := time.Now()
	type testStruct struct {
		ID int32
	}

	sq := shardqueue.NewShardQueue(numShard, queueSize)
	sq.Start(func(msg interface{}) error {
		if v, ok := msg.(testStruct); ok {
			if processCount == totalMsg-1 {
				log.Println("process id", v.ID, "processCount", processCount, "in", time.Since(begin))
			}
			atomic.AddInt32(&processCount, 1)
		}
		return nil
	})

	test := testStruct{}
	for i := range totalMsg {
		test.ID = i
		sq.Shard(strconv.Itoa(int(test.ID)), test)
	}

	sq.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("done")
}
