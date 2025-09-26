package utils

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
)

func GenerateTaskID() string {
	return uuid.New().String()
}

func GenerateUniqID(bucket, suffix string) string {
	mask := fmt.Sprintf("%s:%s", bucket, suffix)
	suffix = fmt.Sprintf("%x", md5.Sum([]byte(mask)))
	return suffix
}
