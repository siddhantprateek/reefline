package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type RedisQueue struct {
	client    *asynq.Client
	server    *asynq.Server
	mux       *asynq.ServeMux
	inspector *asynq.Inspector
	addr      string
}

func NewRedisQueue(addr string, password string) *RedisQueue {
	redisOpt := asynq.RedisClientOpt{
		Addr:     addr,
		Password: password,
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Error processing task %s: %v", task.Type(), err)
			}),
		},
	)

	return &RedisQueue{
		client:    client,
		server:    server,
		mux:       asynq.NewServeMux(),
		inspector: inspector,
		addr:      addr,
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, jobType string, payload interface{}, opts ...Option) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	task := asynq.NewTask(jobType, data)

	var asynqOpts []asynq.Option
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.Delay > 0 {
		asynqOpts = append(asynqOpts, asynq.ProcessIn(options.Delay))
	}
	// Default to 'default' queue and retain completed tasks for 24h
	asynqOpts = append(asynqOpts, asynq.Queue("default"), asynq.Retention(24*time.Hour))

	info, err := q.client.EnqueueContext(ctx, task, asynqOpts...)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

func (q *RedisQueue) RegisterHandler(jobType string, handler func(ctx context.Context, payload []byte) error) {
	q.mux.HandleFunc(jobType, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, t.Payload())
	})
}

func (q *RedisQueue) Start() error {
	go func() {
		if err := q.server.Run(q.mux); err != nil {
			log.Printf("Could not start queue server: %v", err)
		}
	}()
	log.Printf("Redis queue started at %s", q.addr)
	return nil
}

func (q *RedisQueue) Stop() {
	q.client.Close()
	q.inspector.Close()
	q.server.Stop()
	q.server.Shutdown()
	log.Println("Redis queue stopped")
}

func (q *RedisQueue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	taskInfo, err := q.inspector.GetTaskInfo("default", jobID)
	if err != nil {
		return "", err
	}
	return taskInfo.State.String(), nil
}
