package utils

import (
	"crypto/md5"
	"fmt"
)

func GenerateUniqID(bucket, suffix string) string {
	mask := fmt.Sprintf("%s:%s", bucket, suffix)
	suffix = fmt.Sprintf("%x", md5.Sum([]byte(mask)))
	return suffix
}

func ConstructUniqID(bucket, taskID string) string {
	return fmt.Sprintf("watchtower:%s:%s", bucket, taskID)
}
