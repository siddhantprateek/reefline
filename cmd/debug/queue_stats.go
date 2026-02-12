package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hibiken/asynq"
)

func main() {
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	if redisAddr == ":" {
		redisAddr = "localhost:6379"
	}

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	queues, err := inspector.Queues()
	if err != nil {
		log.Fatalf("Could not get queues: %v", err)
	}

	for _, q := range queues {
		fmt.Printf("Queue: %s\n", q)
		info, err := inspector.GetQueueInfo(q)
		if err != nil {
			log.Printf("Error getting info for queue %s: %v", q, err)
			continue
		}
		fmt.Printf("  Size: %d\n", info.Size)
		fmt.Printf("  Active: %d\n", info.Active)
		fmt.Printf("  Pending: %d\n", info.Pending)
		fmt.Printf("  Scheduled: %d\n", info.Scheduled)
		fmt.Printf("  Retry: %d\n", info.Retry)
		fmt.Printf("  Archived: %d\n", info.Archived)
		fmt.Printf("  Completed: %d\n", info.Completed)
		fmt.Printf("  Paused: %v\n", info.Paused)
	}

}
