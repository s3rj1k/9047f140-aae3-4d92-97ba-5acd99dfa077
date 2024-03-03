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

type BackendDef struct {
	MinioClient *minio.Client
	Name        string
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

func (b *Backends) GetMembers() []BackendDef {
	members := b.ch.GetMembers()
	out := make([]BackendDef, 0, len(members))

	for _, el := range members {
		out = append(out, BackendDef{
			Name:        el.String(),
			MinioClient: b.backends[el.String()],
		})
	}

	return out
}
