package process

import "watchtower/internal/domain/core/cloud"

type CreateTaskParams struct {
	BucketID cloud.BucketID
	ObjectID cloud.ObjectID
}
