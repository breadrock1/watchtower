package dto

import "time"

type DocumentObject struct {
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int       `json:"file_size"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}
