package cloud

import "time"

type BucketID = string

type Bucket struct {
	Name      string
	Path      string
	CreatedAt time.Time
}
