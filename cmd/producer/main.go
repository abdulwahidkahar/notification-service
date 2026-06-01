package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	"github.com/abdulwahidkahar/notification-service/internal/queue"
	"github.com/google/uuid"
)

const totalMessages = 2

func main() {
	ctx := context.Background()
	rdb := queue.NewRedisClient()
	defer rdb.Close()

	if err := queue.Ping(ctx, rdb); err != nil {
		log.Fatalf("gagal konek ke Redis: %v", err)
	}

	log.Printf("mulai publish %d messages...", totalMessages)
	start := time.Now()

	for i := 0; i < totalMessages; i++ {
		event := model.NotificationEvent{
			EventID:   uuid.NewString(),
			Type:      model.EventTransferSuccess,
			UserID:    fmt.Sprintf("user_%d", i),
			Amount:    int64(i * 1000),
			CreatedAt: time.Now(),
		}

		if _, err := queue.Publish(ctx, rdb, event); err != nil {
			log.Printf("gagal publish: %v", err)
		}
	}

	elapsed := time.Since(start)
	log.Printf("selesai publish %d messages dalam %v", totalMessages, elapsed)
}
