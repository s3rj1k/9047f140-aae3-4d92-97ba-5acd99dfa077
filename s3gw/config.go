package s3gw

import (
	"sync"

	"github.com/buraksezer/consistent"
	"github.com/minio/minio-go/v7"
)

type Backends struct {
	ch       *consistent.Consistent
	backends map[string]*minio.Client

	mu sync.RWMutex
}

func (b *Backends) Locate(id string) (*minio.Client, string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	backendID := b.ch.LocateKey([]byte(id)).String()

	client, ok := b.backends[backendID]
	if !ok {
		return nil, ""
	}

	return client, backendID
}
