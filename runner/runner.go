package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"

	"github.com/zanderhavgaard/aqueduct/settings"
)

// task to run in a container
type Task struct {
	Name    string
	Type    string
	Command string
	Uses    string
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
	err = c.pullDockerImage()

	containerName := c.Name
	containerConfig := container.Config{
		Image: c.Image,
		// Cmd:   []string{"cat", "/etc/os-release"},
		Cmd: []string{"tail", "-f", "/dev/null"},
	}

	// check if there is another container with the same name
	containerNameIsFree, err := c.checkContainerNameIsFree(ctx, dockerClient)
	if settings.Global.RemoveConflictingContainers {
		if !containerNameIsFree {
			color.Magenta(fmt.Sprintf("Removing container with conflicting name %s ...", c.Name))
			err = c.StopAndRemoveByName(ctx, dockerClient)
			if err != nil {
				panic(err)
			}
			color.Green("Removed conflicting container.")
		}
	}

	// create container
	containerResponse, err := dockerClient.ContainerCreate(ctx, &containerConfig, nil, nil, nil, containerName)
	if err != nil {
		panic(err)
	}
	containerID := containerResponse.ID
	fmt.Println("Created container with ID:", containerID)

	fmt.Println("Starting container ...")
	// actually start the container
	err = dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	// how many tasks to run in this container
	num_tasks := len(c.Tasks)

	color.Magenta("Executing tasks ...")

	for index, task := range c.Tasks {

		color.Blue("--- Setting up task")
		color.Blue(fmt.Sprintf("Execting task %d of %d", index+1, num_tasks))
		color.Blue(fmt.Sprintf("Task name: %s", task.Name))

		returnCode, err := task.execute(ctx, dockerClient, containerResponse)
		if err != nil {
			panic(err)
		}

		fmt.Println(returnCode)
		color.Green(fmt.Sprintf("Done execting task: %s", task.Name))
	}

	// define the timeout for stopping the containers
	// var timeout time.Duration = time.Second * 30
	var timeout time.Duration

	if settings.Global.GracefulContainerShutdown {
		// stop containers, allowing for graceful shutdown
		err = dockerClient.ContainerStop(ctx, containerID, nil)
	} else {
		// stop container immediately
		timeout = time.Second * 0
		err = dockerClient.ContainerStop(ctx, containerID, &timeout)
	}
	if err != nil {
		panic(err)
	}

	//  remove containers
	if settings.Global.RemoveContainers {
		err = dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (t Task) execute(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {

	if t.Type == "shell" {
		return t.executeShellCommand(ctx, dockerClient, containerResponse)
	} else if t.Type == "github-action" {
		return t.executeGithubAction(ctx, dockerClient, containerResponse)
	} else {
		panic("Task type not implemented")
	}
}

// execute tasks in containers for the run
func ExecuteRun(run Run, mode string) error {

	fmt.Println()
	color.Magenta("Executing Containers ...")
	fmt.Println()

	// number of containers in this run
	num_containers := len(run.Containers)

	// loop over containers and execute their tasks
	if mode == "all" {
		for index, container := range run.Containers {

			color.Magenta("--- Setting up container")
			color.Magenta(fmt.Sprintf("Executing container %d of %d", index+1, num_containers))
			color.Magenta(fmt.Sprintf("Container name: %s", container.Name))

			container.executeTasks()

			color.Green("Done executing container.")
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println("--------------------")
	color.Green("Done executing all containers.")
	fmt.Println()

	return nil
}
