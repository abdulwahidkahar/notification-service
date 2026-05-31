package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

const ConsumerGroup = "notification-workers"

func CreateConsumerGroup(ctx context.Context, rdb *redis.Client) error {
	// cek apakah stream sudah ada
	exists, _ := rdb.Exists(ctx, StreamName).Result()
	if exists == 0 {
		// buat stream kosong dengan XADD lalu langsung hapus
		id, _ := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: StreamName,
			ID:     "*",
			Values: map[string]interface{}{"init": "1"},
		}).Result()
		rdb.XDel(ctx, StreamName, id)
	}

	err := rdb.XGroupCreate(ctx, StreamName, ConsumerGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("gagal buat consumer group: %w", err)
	}
	return nil
}

func Consume(ctx context.Context, rdb *redis.Client, consumerName string) ([]redis.XMessage, error) {
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumerName,
		Streams:  []string{StreamName, ">"},
		Count:    10,
		Block:    2000,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []redis.XMessage{}, nil
		}

		if strings.Contains(err.Error(), "i/o timeout") {
			return []redis.XMessage{}, nil
		}
		return nil, fmt.Errorf("gagal consume: %w", err)
	}

	return streams[0].Messages, nil
}

func Acknowledge(ctx context.Context, rdb *redis.Client, messageID string) error {
	return rdb.XAck(ctx, StreamName, ConsumerGroup, messageID).Err()
}
