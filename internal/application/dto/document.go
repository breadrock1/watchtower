package dto

import "time"

type StorageDocument struct {
	Content    string         `json:"content"`
	SSDEEP     string         `json:"ssdeep"`
	Class      string         `json:"class"`
	FileName   string         `json:"file_name"`
	FilePath   string         `json:"file_path"`
	FileSize   int            `json:"file_size"`
	CreatedAt  time.Time      `json:"created_at"`
	ModifiedAt time.Time      `json:"modified_at"`
	Tokens     ComputedTokens `json:"embeddings"`
}
