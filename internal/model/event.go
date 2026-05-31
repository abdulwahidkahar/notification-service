package model

import "time"

type EventType string

const (
	EventTransferSuccess EventType = "transfer.success"
	EventTopUpConfirmed  EventType = "topup.confirmed"
)

const MaxRetry = 3

type NotificationEvent struct {
	EventID    string    `json:"event_id"`
	Type       EventType `json:"type"`
	UserID     string    `json:"user_id"`
	Amount     int64     `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	RetryCount int       `json:"retry_count"`
	LastError  string    `json:"last_error"`
}
