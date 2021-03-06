package taskqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskQueue struct {
	redis             Redis
	taskQueueKey      string
	inProgressTaskKey string
	consumeScriptSha  string
	maxRetries        int
	operationTimeout  time.Duration
}

type Task struct {
	ID         uuid.UUID
	Payload    interface{}
	RetryCount int
	Wait       time.Duration
}

type ConsumeFunc func(context.Context, uuid.UUID, interface{}) error

func (t *Task) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal *Task: %w", err)
	}

	return data, nil
}

func NewTaskQueue(ctx context.Context, redisClient Redis, options *Options) (*TaskQueue, error) {
	options.setDefaults()

	consumeScriptBytes, err := os.ReadFile(getConsumeScriptPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read consume script file: %w", err)
	}

	consumeScriptSHA, err := redisClient.ScriptLoad(ctx, string(consumeScriptBytes)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load consume script file into redis: %w", err)
	}

	taskQueue := &TaskQueue{
		redis:        redisClient,
		taskQueueKey: fmt.Sprintf("taskqueue:%s:tasks:%s", options.Namespace, options.QueueKey),
		inProgressTaskKey: fmt.Sprintf(
			"taskqueue:%s:workers:%s:tasks:%s",
			options.Namespace, options.WorkerID, options.QueueKey,
		),
		consumeScriptSha: consumeScriptSHA,
		maxRetries:       options.MaxRetries,
		operationTimeout: options.OperationTimeout,
	}

	return taskQueue, nil
}

func NewDefaultRedis(options *Options) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: options.StorageAddress}) //nolint:exhaustruct
}

func (t *TaskQueue) ProduceAt(ctx context.Context, payload interface{}, executeAt time.Time) (uuid.UUID, error) {
	logger := newLogger()

	task := &Task{
		ID:         uuid.New(),
		Payload:    payload,
		RetryCount: 0,
		Wait:       0,
	}

	logger.Debugf("producing task %s %v", task.ID, task.Payload)

	err := t.produceAt(ctx, task, executeAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to produce task: %w", err)
	}

	return task.ID, nil
}

func (t *TaskQueue) Consume(
	ctx context.Context,
	consume ConsumeFunc,
) {
	var (
		ticker = time.NewTicker(time.Second)
		logger = newLogger().WithFields(logrus.Fields{
			"operation": "consumer",
		})
	)

	for {
		select {
		case <-ticker.C:
			logger.Info("consuming task")

			if err := t.consume(ctx, consume); err != nil {
				logger.WithError(err).Error("failed to call consume function")
			}

		case <-ctx.Done():
			logger.Info("stopping")

			return
		}
	}
}

func getConsumeScriptPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	return filepath.Join(basepath, "consume.lua")
}

func (t *TaskQueue) consume(
	ctx context.Context,
	consume ConsumeFunc,
) error {
	logger := newLogger()

	task, err := t.getTask(ctx)
	if errors.Is(err, ErrNoTaskToConsume) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	logger = withTaskLabels(logger, task)
	defer t.removeInProgressTask(ctx, task)

	logger.Debug("consuming task")

	err = consume(ctx, task.ID, task.Payload)
	if err != nil {
		logger.WithError(err).Debug("failed to consume, retrying after backoff")

		retryErr := t.produceRetry(ctx, task)
		if retryErr != nil {
			return ErrTaskLost(
				task.ID,
				fmt.Errorf("failed to run consume func and failed to enqueue it for retry: %w", retryErr),
			)
		}

		return fmt.Errorf("failed to run consume func: %w", err)
	}

	logger.Debug("successfully consumed task")

	return nil
}

func (t *TaskQueue) getTask(ctx context.Context) (*Task, error) {
	logger := newLogger()

	now := time.Now()
	keys := []string{
		t.taskQueueKey,
		t.inProgressTaskKey,
	}
	args := []interface{}{
		fmt.Sprintf("%d", now.Unix()),
	}

	logger.Debugf("fetching task to execute")

	taskInterface, err := t.redis.EvalSha(ctx, t.consumeScriptSha, keys, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute consume script: %w", err)
	}

	if taskInterface == StatusOK {
		logger.Debug("no tasks to execute")

		return nil, ErrNoTaskToConsume
	}

	taskStr, ok := taskInterface.(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast task: %w", ErrInvalidTaskType)
	}

	task := new(Task)

	err = json.Unmarshal([]byte(taskStr), task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task and permanently lost it: %w", err)
	}

	return task, nil
}

func (t *TaskQueue) produceAt(
	ctx context.Context,
	task *Task,
	executeAt time.Time,
) error {
	if err := t.redis.ZAdd(ctx, t.taskQueueKey, &redis.Z{
		Score:  float64(executeAt.Unix()),
		Member: task,
	}).Err(); err != nil {
		return fmt.Errorf("failed to zadd message: %w", err)
	}

	return nil
}

func (t *TaskQueue) produceRetry(ctx context.Context, task *Task) error {
	now := time.Now()

	if t.maxRetries >= 0 && task.RetryCount >= t.maxRetries {
		return fmt.Errorf("%w: %s, %d", ErrMaxTaskReties, task.ID, task.RetryCount)
	}

	wait := time.Second
	if task.Wait > 0 {
		wait = 2 * task.Wait // nolint:gomnd
	}

	retryTask := &Task{
		ID:         task.ID,
		Payload:    task.Payload,
		RetryCount: task.RetryCount + 1,
		Wait:       wait,
	}

	executeAt := now.Add(wait)

	err := t.produceAt(ctx, retryTask, executeAt)
	if err != nil {
		return fmt.Errorf("failed to produce retry task: %w", err)
	}

	return nil
}

func (t *TaskQueue) removeInProgressTask(ctx context.Context, task *Task) {
	logger := newLogger()
	logger = withTaskLabels(logger, task)

	err := t.redis.Del(ctx, t.inProgressTaskKey).Err()
	if err != nil {
		logger.
			WithError(err).
			Warn("failed to delete worker in progress task, might duplicate if worker restart now")
	}
}
