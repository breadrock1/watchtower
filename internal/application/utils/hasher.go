package utils

import (
	"crypto/md5"
	"fmt"
)

func GenerateUniqID(bucket, suffix string) string {
	if suffix == "*" {
		return fmt.Sprintf("watchtower:%s:%s", bucket, suffix)
	}

	suffix = fmt.Sprintf("%x", md5.Sum([]byte(suffix)))
	return fmt.Sprintf("watchtower:%s:%s", bucket, suffix)
}
