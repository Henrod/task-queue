package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Henrod/task-queue/taskqueue"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Payload struct {
	Body string
}

func handleStop(cancel context.CancelFunc) {
	logger := logrus.New()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	logger.Info("received termination signal, waiting for operations to finish")
	cancel()
}

const (
	typeConsumer = "consumer"
	typeProducer = "producer"
)

func runConsumer(ctx context.Context, taskQueue *taskqueue.TaskQueue) {
	logger := logrus.New().WithFields(logrus.Fields{
		"operation": "consumer",
	})

	logger.Info("consuming task")
	taskQueue.Consume(
		ctx,
		func(ctx context.Context, taskID uuid.UUID, payload interface{}) error {
			logger.Printf("consumed task %s: %v\n", taskID, payload)

			return nil
		},
	)
}

func runProducer(ctx context.Context, taskQueue *taskqueue.TaskQueue) {
	var (
		ticker = time.NewTicker(time.Second)
		logger = logrus.New().WithFields(logrus.Fields{
			"operation": "producer",
		})
	)

	id := 0

	for {
		select {
		case <-ticker.C:
			logger.Info("producing task")

			taskID, err := taskQueue.ProduceAt(ctx, &Payload{Body: fmt.Sprintf("%d", id)}, time.Now())
			if err != nil {
				logger.WithError(err).Error("failed to enqueue task")

				break
			}

			id++

			logger.Infof("enqueued task %s", taskID)

		case <-ctx.Done():
			logger.Info("stopping")

			return
		}
	}
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	serverType := os.Args[1]

	ctx, cancel := context.WithCancel(context.Background())

	go handleStop(cancel)

	options := &taskqueue.Options{
		QueueKey:         "dummy-consumer",
		Namespace:        "simple",
		StorageAddress:   "localhost:6379",
		WorkerID:         "worker1",
		MaxRetries:       -1,
		OperationTimeout: time.Minute,
	}

	taskQueue, err := taskqueue.NewTaskQueue(ctx, taskqueue.NewDefaultRedis(options), options)
	if err != nil {
		panic(err)
	}

	switch serverType {
	case typeConsumer:
		runConsumer(ctx, taskQueue)
	case typeProducer:
		runProducer(ctx, taskQueue)
	}
}
