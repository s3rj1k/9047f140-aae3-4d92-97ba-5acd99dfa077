package s3gw

import (
	"github.com/cespare/xxhash" // just a quick hash function for consistent hash module
)

const (
	ConsistentHashPartitionCountEnvKey    = "CONSISTENT_HASH_PARTITION_COUNT"
	ConsistentHashReplicationFactorEnvKey = "CONSISTENT_HASH_REPLICATION_FACTOR"
	ConsistentHashLoadEnvKey              = "CONSISTENT_HASH_LOAD"
)

type Member string

func (m Member) String() string {
	return string(m)
}

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}
