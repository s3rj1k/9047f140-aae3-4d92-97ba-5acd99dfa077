package s3gw

import (
	"os"
	"strconv"
)

func MustGetIntFromEnv(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}

	return n
}

func MustGetFloat64FromEnv(key string) float64 {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}

	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(err)
	}

	return n
}
