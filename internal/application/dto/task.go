package dto

import (
	"time"

	"github.com/google/uuid"
)

type TaskEvent struct {
	Id         uuid.UUID `json:"id"`
	Bucket     string    `json:"bucket"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	Status     int       `json:"status"`
	StatusText string    `json:"status_text"`
}
