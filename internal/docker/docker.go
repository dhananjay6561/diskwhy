package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
)

// Result holds Docker resource sizes broken down by resource type.
type Result struct {
	UnusedImageBytes int64
	UsedImageBytes   int64
	VolumeBytes      int64
	UnusedImageCount int
	UsedImageCount   int
	VolumeCount      int
	SocketPath       string
}

// discoverDockerSocket tries each socket location in the order specified by
// PRD §5.3.1 and returns the first path that exists on disk.
func discoverDockerSocket(home string) (string, error) {
	candidates := []string{
		os.Getenv("DOCKER_HOST"),
		"/var/run/docker.sock",
		filepath.Join(home, ".docker", "run", "docker.sock"),
		fmt.Sprintf("/run/user/%d/docker.sock", os.Getuid()),
		filepath.Join(home, ".colima", "default", "docker.sock"),
	}

	var tried []string
	for _, path := range candidates {
		if path == "" {
			continue
		}
		path = strings.TrimPrefix(path, "unix://")
		tried = append(tried, path)
		if _, err := os.Lstat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no Docker socket found\nFix: start Docker Desktop or colima, or set DOCKER_HOST to one of: %s",
		strings.Join(tried, ", "))
}

// Query connects to Docker, lists images/containers/volumes, and returns
// resource sizes. Returns an empty Result (no error) when Docker is unavailable.
func Query(ctx context.Context, home string, verbose bool) (*Result, error) {
	socketPath, err := discoverDockerSocket(home)
	if err != nil {
		return &Result{}, nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  docker: using socket %s\n", socketPath)
	}

	queryCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost("unix://"+socketPath),
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return &Result{}, nil
	}
	defer cli.Close()

	containers, err := cli.ContainerList(queryCtx, container.ListOptions{All: true})
	if err != nil {
		return &Result{}, nil
	}
	usedImageIDs := make(map[string]struct{}, len(containers))
	for _, c := range containers {
		usedImageIDs[c.ImageID] = struct{}{}
	}

	images, err := cli.ImageList(queryCtx, image.ListOptions{})
	if err != nil {
		return &Result{}, nil
	}

	res := &Result{SocketPath: socketPath}
	for _, img := range images {
		if _, used := usedImageIDs[img.ID]; used {
			res.UsedImageBytes += img.Size
			res.UsedImageCount++
		} else {
			res.UnusedImageBytes += img.Size
			res.UnusedImageCount++
		}
	}

	vols, err := cli.VolumeList(queryCtx, volume.ListOptions{})
	if err != nil {
		return res, nil
	}
	for _, v := range vols.Volumes {
		if v.UsageData != nil && v.UsageData.Size >= 0 {
			res.VolumeBytes += v.UsageData.Size
			res.VolumeCount++
		}
	}

	return res, nil
}
