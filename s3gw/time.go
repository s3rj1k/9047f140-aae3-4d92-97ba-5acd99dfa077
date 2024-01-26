package s3gw

import (
	"time"
)

func MustParseDuration(s string) time.Duration {
	td, _ := time.ParseDuration(s)
	return td
}
