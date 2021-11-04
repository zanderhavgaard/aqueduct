package runner

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// task to run in a container
type Task struct {
	Name    string
	Command string
}

// a container to run tasks in
type Container struct {
	Name  string
	Image string
	Tasks []Task
}

// a run of a pipeline
type Run struct {
	Name       string
	Containers []Container
}

func (c Container) executeTasks() error {
	// setup context
	ctx := context.Background()
	// get a docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// pull the image before starting container
	out, err := dockerClient.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)

	containerName := c.Name
	containerConfig := container.Config{
		Image: c.Image,
		Cmd:   []string{"cat", "/etc/os-release"},
	}

	// create container
	response, err := dockerClient.ContainerCreate(ctx, &containerConfig, nil, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	// actually start the container
	err = dockerClient.ContainerStart(ctx, response.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Running Job", c.Name, "in Container:", response.ID)

	statusCh, errCh := dockerClient.ContainerWait(ctx, response.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case status := <-statusCh:
		fmt.Println("Container status:", status)
	}

	return nil
}

// execute tasks in containers for the run
func ExecuteRun(run Run, mode string) error {

	fmt.Println("---")
	fmt.Println(run)
	fmt.Println("---")

	if mode == "all" {
		fmt.Println("Running all tasks in run ...")

		for _, container := range run.Containers {
			fmt.Println(container.Name)
			container.executeTasks()
		}
	}

	return nil
}
