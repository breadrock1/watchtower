package utils

import (
	"crypto/md5"
	"fmt"

	"github.com/glaslos/ssdeep"
)

func ComputeMd5(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func ComputeSSDEEP(data []byte) (string, error) {
	hashData, err := ssdeep.FuzzyBytes(data)
	if err != nil {
		return "", err
	}

	return hashData, nil
}
