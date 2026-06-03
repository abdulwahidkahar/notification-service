package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdulwahidkahar/notification-service/internal/model"
	"github.com/abdulwahidkahar/notification-service/internal/notification"
	"github.com/abdulwahidkahar/notification-service/internal/queue"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	emailCfg, err := notification.NewEmailConfig()
	if err != nil {
		log.Fatalf("gagal init email config: %v", err)
	}

	rdb := queue.NewRedisClient()
	defer rdb.Close()

	if err := queue.Ping(ctx, rdb); err != nil {
		log.Fatalf("gagal konek ke Redis: %v", err)
	}

	if err := queue.CreateConsumerGroup(ctx, rdb); err != nil {
		log.Fatalf("gagal buat consumer group: %v", err)
	}

	consumerName := fmt.Sprintf("worker-%d", os.Getpid())
	log.Printf("worker started — name: %s", consumerName)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutdown signal received...")
		cancel()
	}()

	var totalProcessed int
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker stopped gracefully")
			return
		default:
			messages, err := queue.Consume(ctx, rdb, consumerName)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("error consume: %v", err)
				continue
			}

			for _, msg := range messages {
				err := process(ctx, rdb, msg, emailCfg)
				if err != nil {
					payload, _ := msg.Values["payload"].(string)
					var event model.NotificationEvent
					json.Unmarshal([]byte(payload), &event)

					event.RetryCount++
					event.LastError = err.Error()

					if event.RetryCount >= model.MaxRetry {
						if dlErr := queue.PublishDeadLetter(ctx, rdb, event); dlErr != nil {
							log.Printf("gagal publish ke dead letter: %v", dlErr)
						} else {
							log.Printf("[DEAD LETTER] message dipindah setelah %d retry — user: %s | error: %s",
								event.RetryCount, event.UserID, event.LastError)
						}
					} else {
						if _, retryErr := queue.Publish(ctx, rdb, event); retryErr != nil {
							log.Printf("gagal retry: %v", retryErr)
						} else {
							log.Printf("[RETRY %d/%d] — user: %s | error: %s",
								event.RetryCount, model.MaxRetry, event.UserID, event.LastError)
						}
					}
				}

				if err := queue.Acknowledge(ctx, rdb, msg.ID); err != nil {
					log.Printf("gagal ack message %s: %v", msg.ID, err)
				}

				totalProcessed++
				if totalProcessed%100 == 0 {
					elapsed := time.Since(start)
					log.Printf("[STATS] processed: %d | elapsed: %v | throughput: %.0f msg/s",
						totalProcessed,
						elapsed,
						float64(totalProcessed)/elapsed.Seconds(),
					)
				}
			}
		}
	}
}

func process(ctx context.Context, rdb *redis.Client, msg redis.XMessage, emailCfg *notification.EmailConfig) error {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("invalid payload format")
	}

	var event model.NotificationEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return fmt.Errorf("gagal unmarshal: %w", err)
	}

	to := os.Getenv("SMTP_TO")

	var subject, body string
	switch event.Type {
	case model.EventTransferSuccess:
		subject = "Transfer Berhasil"
		body = fmt.Sprintf("<h2>Transfer Berhasil</h2><p>User <b>%s</b> melakukan transfer sebesar <b>Rp%d</b>.</p>", event.UserID, event.Amount)
	case model.EventTopUpConfirmed:
		subject = "Top Up Dikonfirmasi"
		body = fmt.Sprintf("<h2>Top Up Dikonfirmasi</h2><p>User <b>%s</b> melakukan top up sebesar <b>Rp%d</b>.</p>", event.UserID, event.Amount)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	if err := emailCfg.Send(to, subject, body); err != nil {
		return err
	}

	log.Printf("[EMAIL SENT] %s → %s | user: %s", event.Type, to, event.UserID)
	return nil
}
