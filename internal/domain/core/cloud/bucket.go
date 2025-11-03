package cloud

import "time"

type BucketID = string

type Bucket struct {
	ID        BucketID
	Path      string
	CreatedAt time.Time
}
