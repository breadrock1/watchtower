package domain

import "time"

type Document struct {
	Index      string
	FileName   string
	FilePath   string
	FileSize   int
	Content    string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
