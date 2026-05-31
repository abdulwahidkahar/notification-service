package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const StreamName = "notification:events"

func Publish(ctx context.Context, rdb *redis.Client, event model.NotificationEvent) (string, error) {

	payload, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("gagal marshal event: %w", err)
	}

	id, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{
			"payload": payload,
		},
	}).Result()

	return id, nil
}

const DeadLetterStream = "notification:dead-letter"

func PublishDeadLetter(ctx context.Context, rdb *redis.Client, event model.NotificationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("gagal marshal event: %w", err)
	}

	_, err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: DeadLetterStream,
		Values: map[string]interface{}{
			"payload": payload,
		},
	}).Result()

	return err
}
