package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	"github.com/abdulwahidkahar/notification-service/internal/queue"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	conn, err := queue.NewRabbitMQConn()
	if err != nil {
		log.Fatalf("gagal konek RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("gagal buat channel: %v", err)
	}
	defer ch.Close()

	ctx := context.Background()

	events := []model.NotificationEvent{
		{
			EventID:   uuid.NewString(),
			Type:      model.EventTransferSuccess,
			UserID:    "user_123",
			Amount:    150000,
			CreatedAt: time.Now(),
		},
		{
			EventID:   uuid.NewString(),
			Type:      model.EventTopUpConfirmed,
			UserID:    "user_456",
			Amount:    500000,
			CreatedAt: time.Now(),
		},
	}

	for _, event := range events {
		if err := queue.PublishToRabbitMQ(ctx, ch, event); err != nil {
			log.Printf("gagal publish: %v", err)
			continue
		}
		fmt.Printf("published — type: %s | user: %s\n", event.Type, event.UserID)
	}
}
