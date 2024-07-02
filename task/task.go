package task

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	Cmd           []string
	Image         string
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

type Task struct {
	ID            uuid.UUID
	Name          string
	State         State
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	portBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Docker struct {
	Client      *client.Client
	Config      Config
	ContainerId string
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	// Pull image
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	reader, err := cli.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, reader)

	// rp := container.RestartPolicy{
	// 	Name: d.Config.RestartPolicy,
	// }

	// r := container.Resources{
	// 	Memory: d.Config.Memory,
	// }

	cc := container.Config{
		Image: d.Config.Image,
		Env:   d.Config.Env,
	}

	// hc := container.HostConfig{
	// 	RestartPolicy:   rp,
	// 	Resources:       r,
	// 	PublishAllPorts: true,
	// }

	resp, err := cli.ContainerCreate(
		ctx, &cc, nil, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf(
			"Error creating container using image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	// start container
	// not able to define
	err2 := cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err2 != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err2)
		return DockerResult{Error: err2}
	}

	d.ContainerId = resp.ID

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{ContainerId: resp.ID, Action: "start", Result: "success"}
}

// func (cli *Client) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) container.ContainerCreateCreatedBody

func (d *Docker) Stop() DockerResult {
	ctx := context.Background()
	log.Printf("Attempting to stop container %v", d.ContainerId)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	err2 := cli.ContainerStop(ctx, d.ContainerId, container.StopOptions{})
	if err2 != nil {
		panic(err2)
	}

	removeOptions := container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	}

	err = d.Client.ContainerRemove(
		ctx,
		d.ContainerId,
		removeOptions,
	)
	if err != nil {
		panic(err)
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}
