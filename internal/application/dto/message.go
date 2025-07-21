package dto

import "github.com/google/uuid"

type Message struct {
	EventId uuid.UUID `json:"event_id"`
	Body    TaskEvent `json:"body"`
}
