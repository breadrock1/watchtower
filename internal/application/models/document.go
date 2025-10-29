package models

import "time"

type Document struct {
	Index      string
	Name       string
	Path       string
	Size       int
	Content    string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
