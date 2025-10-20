package domain

import "time"

type Document struct {
	FileName   string
	FilePath   string
	FileSize   int
	Content    string
	Class      string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
