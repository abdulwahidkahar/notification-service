package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const RabbitMQQueue = "notification.events"

func NewRabbitMQConn() (*amqp.Connection, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("gagal konek ke RabbitMQ: %w", err)
	}

	return conn, nil
}

func PublishToRabbitMQ(ctx context.Context, ch *amqp.Channel, event model.NotificationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("gagal marshal event: %w", err)
	}

	_, err = ch.QueueDeclare(
		RabbitMQQueue,
		true,  // durable — queue tetap ada meski RabbitMQ restart
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("gagal declare queue: %w", err)
	}

	err = ch.PublishWithContext(ctx,
		"",            // exchange
		RabbitMQQueue, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
			DeliveryMode: amqp.Persistent, // message tetap ada meski RabbitMQ restart
		},
	)

	if err != nil {
		return fmt.Errorf("gagal publish ke RabbitMQ: %w", err)
	}

	log.Printf("event published to RabbitMQ — type: %s | user: %s", event.Type, event.UserID)
	return nil
}

func ConsumeFromRabbitMQ(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	_, err := ch.QueueDeclare(
		RabbitMQQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal declare queue: %w", err)
	}

	// prefetch 1 — worker hanya ambil 1 message, baru ambil lagi setelah selesai
	ch.Qos(1, 0, false)

	msgs, err := ch.Consume(
		RabbitMQQueue,
		"",    // consumer name — auto generate
		false, // auto-ack — kita manual ACK
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal consume dari RabbitMQ: %w", err)
	}

	return msgs, nil
}
