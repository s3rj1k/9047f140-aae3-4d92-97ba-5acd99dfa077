package s3gw

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	S3ContainerNamePatternEnvKey = "S3_CONTAINER_NAME_PATTERN"
	S3APIPortEnvKey              = "S3_API_PORT"
)

func ListInspectRunningContainersFilteredByName(ctx context.Context, cli *client.Client, keep func(string) bool) ([]types.ContainerJSON, error) {
	containers, err := cli.ContainerList(ctx,
		types.ContainerListOptions{
			All: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var filteredInspectedContainers []types.ContainerJSON

	for _, container := range containers {
		for _, name := range container.Names {
			if !keep(name) {
				continue
			}

			inspect, err := cli.ContainerInspect(context.Background(), container.ID)
			if err != nil {
				return nil, fmt.Errorf("error inspecting container with ID %s: %w", container.ID, err)
			}

			filteredInspectedContainers = append(filteredInspectedContainers, inspect)
			break
		}
	}

	sort.Slice(filteredInspectedContainers, func(i, j int) bool {
		return filteredInspectedContainers[i].Name < filteredInspectedContainers[j].Name
	})

	return filteredInspectedContainers, nil
}

func TryPing(ctx context.Context, cli *client.Client, attempts int) (finalError error) { // to verify that backend connection is stable
	wait := 1 * time.Second

	fmt.Printf("Waiting for stable Docker backend connection...\n")

	for i := 0; i < attempts; i++ {
		_, err := cli.Ping(ctx)
		if err != nil {
			fmt.Printf("Ping attempt failed, error: %v\n", err)

			if i < attempts-1 { // increase wait time only if more attempts are left
				wait *= 2
			}

			finalError = err
		}

		if i < attempts-1 { // wait before the next attempt, except after the last one
			time.Sleep(wait)
		}
	}

	return
}

func GetBackendContainerIDs(ctx context.Context, cli *client.Client) ([]string, error) {
	err := TryPing(ctx, cli, 3)
	if err != nil {
		return nil, err
	}

	containers, err := ListInspectRunningContainersFilteredByName(ctx, cli, func(name string) bool {
		return strings.Contains(name, os.Getenv(S3ContainerNamePatternEnvKey))
	})
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(containers))

	for _, el := range containers {
		out = append(out, el.ID)
	}

	return out, nil
}

func GetBackendAddresses(ctx context.Context, cli *client.Client, containerID string) (ips []net.IP, err error) {
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	networkSettings := containerJSON.NetworkSettings
	if networkSettings == nil {
		return nil, fmt.Errorf("network settings not found for container %s", containerID)
	}

	for _, network := range networkSettings.Networks {
		ip := net.ParseIP(network.IPAddress)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address found for container %s", containerID)
		}

		ips = append(ips, ip)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no valid network addresses found for container %s", containerID)
	}

	sort.Slice(ips, func(i, j int) bool {
		iIsIPv4 := ips[i].To4() != nil
		jIsIPv4 := ips[j].To4() != nil

		if iIsIPv4 == jIsIPv4 {
			return bytes.Compare(ips[i], ips[j]) < 0
		}

		return iIsIPv4
	})

	return ips, nil
}

func GetBackendCredentials(ctx context.Context, cli *client.Client, containerID string) (user, password string, err error) {
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", "", fmt.Errorf("error inspecting container with ID %s: %w", containerID, err)
	}

	for _, envVar := range containerJSON.Config.Env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "MINIO_ACCESS_KEY", "MINIO_ROOT_USER":
			user = parts[1]
		case "MINIO_SECRET_KEY", "MINIO_ROOT_PASSWORD":
			password = parts[1]
		}
	}

	if user == "" || password == "" {
		return "", "", fmt.Errorf("failed to find S3 backend credentials in container %s", containerID)
	}

	return user, password, nil
}
