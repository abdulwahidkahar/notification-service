package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	"github.com/abdulwahidkahar/notification-service/internal/notification"
	"github.com/abdulwahidkahar/notification-service/internal/queue"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	emailCfg, err := notification.NewEmailConfig()
	if err != nil {
		log.Fatalf("gagal init email config: %v", err)
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

	msgs, err := queue.ConsumeFromRabbitMQ(ch)
	if err != nil {
		log.Fatalf("gagal setup consumer: %v", err)
	}

	log.Println("rabbitmq worker started — waiting for messages...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-quit:
			log.Println("worker stopped gracefully")
			return
		case msg := <-msgs:
			var event model.NotificationEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("gagal unmarshal: %v", err)
				msg.Nack(false, false) // buang message, jangan requeue
				continue
			}

			to := os.Getenv("SMTP_TO")
			content, err := notification.BuildEmailContent(event)
			if err != nil {
				log.Print(err)
				msg.Ack(false)
				continue
			}

			if err := emailCfg.Send(to, content.Subject, content.Body); err != nil {
				log.Printf("gagal kirim email: %v", err)
				msg.Nack(false, true) // requeue — coba lagi nanti
				continue
			}

			log.Printf("[EMAIL SENT] %s → %s | user: %s", event.Type, to, event.UserID)
			msg.Ack(false) // ACK — message berhasil diproses
		}
	}
}
