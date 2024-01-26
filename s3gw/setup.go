package s3gw

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Configure(ctx context.Context) (*Backends, error) {
	backendsConfig := new(Backends)

	backendsConfig.backends = make(map[string]*minio.Client)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker instance: %w", err)
	}

	defer cli.Close()

	backends, err := GetBackendContainerIDs(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("failed to discover S3 backends: %w", err)
	}

	if len(backends) == 0 {
		return nil, fmt.Errorf("no S3 backends discovered")
	}

	members := []consistent.Member{}

	for _, el := range backends {
		members = append(members, Member(el))
	}

	consistentHashConfig := consistent.Config{
		PartitionCount:    MustGetIntFromEnv(ConsistentHashPartitionCountEnvKey),
		ReplicationFactor: MustGetIntFromEnv(ConsistentHashReplicationFactorEnvKey),
		Load:              MustGetFloat64FromEnv(ConsistentHashLoadEnvKey),

		Hasher: hasher{},
	}

	backendsConfig.ch = consistent.New(members, consistentHashConfig)

	for _, backend := range backendsConfig.ch.GetMembers() {
		user, password, err := GetBackendCredentials(ctx, cli, backend.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get S3 backend credentials for %s", backend)
		}

		addrs, err := GetBackendAddresses(ctx, cli, backend.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get S3 backend addresses for %s", backend)
		}

		isAlive := false
		cfg := MustCreateNewS3BackendConfig(
			net.JoinHostPort(
				addrs[0].String(), // assuming that first address from Docker Inspect is valid one
				os.Getenv(S3APIPortEnvKey),
			),
			&minio.Options{
				Creds:  credentials.NewStaticV4(user, password, ""),
				Secure: false, // assuming no secure config is set (internal network)
			},
		)

		for i := 1; i <= 5; i++ { // wait for backend to be alive or hard fail
			time.Sleep(MustParseDuration(fmt.Sprintf("%ds", i)))

			err = CheckS3BackendLiveliness(ctx, cfg)
			if err == nil {
				isAlive = true

				break
			}
		}

		if !isAlive {
			return nil, fmt.Errorf("failed to check S3 backend liveliness for %s", backend)
		}

		log.Printf("Using S3 backend with Docker ID: %s", backend)

		backendsConfig.backends[backend.String()] = cfg
	}

	return backendsConfig, nil
}
