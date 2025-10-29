package domain

import "time"

type Bucket struct {
	Name      string
	Path      string
	CreatedAt time.Time
}
