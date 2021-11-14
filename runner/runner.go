package runner

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

	// imagePullOutput, err := dockerClient.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	// if err != nil {
	// panic(err)
	// }
	// io.Copy(os.Stdout, imagePullOutput)

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

	// wait for the container to finish running
	// statusCh, errCh := dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	// select {
	// case err := <-errCh:
	// if err != nil {
	// panic(err)
	// }
	// case status := <-statusCh:
	// fmt.Println("Container status:", status)
	// }

	// print logs from the container
	// logsOutput, err := dockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true})

	// fmt.Println("foo", logsOutput)
	// stdcopy.StdCopy(os.Stdout, os.Stderr, logsOutput)

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

func (c Container) pullDockerImage() error {
	if settings.Global.SkipImagePull {
		color.Green("Skipping docker image pull.")
		return nil
	}
	color.Magenta(fmt.Sprintf("Pulling docker image: %s ...", c.Image))

	// setup context
	ctx := context.Background()
	// get a docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// pull the image before starting container
	imagePullOutput, err := dockerClient.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// io.Copy(os.Stdout, imagePullOutput)
	stdout, err := ioutil.ReadAll(imagePullOutput)
	if err != nil {
		panic(err)
	}

	stdOutAsString := string(stdout)
	if strings.Contains(stdOutAsString, "Image is up to date") {
		color.Green("Image is up-to-date")
		if settings.Global.Verbose {
			fmt.Println("Verbose image pull output:")
			fmt.Println(stdOutAsString)
		}
	} else {
		color.Green("Downloaded newer image.")
		if settings.Global.Verbose {
			fmt.Println("Verbose image pull output:")
			fmt.Println(stdOutAsString)
		}
	}

	return nil
}

func (c Container) checkContainerNameIsFree(ctx context.Context, dockerClient *client.Client) (bool, error) {
	color.Magenta("Checking that the container name is free ...")
	free := true
	options := types.ContainerListOptions{}
	// get list of containers
	containers, err := dockerClient.ContainerList(ctx, options)
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		for _, name := range container.Names {
			// docker adds a '/' in front of the name
			nameNoLeadingSlash := strings.Replace(name, "/", "", 1)
			if nameNoLeadingSlash == c.Name {
				free = false
			}
		}
	}
	return free, nil
}

func (c Container) StopAndRemoveByName(ctx context.Context, dockerClient *client.Client) error {
	fmt.Println("Removing container with name", c.Name)
	// get list of containers
	options := types.ContainerListOptions{}
	containers, err := dockerClient.ContainerList(ctx, options)
	if err != nil {
		panic(err)
	}
	// get the ID of the container with the matching name
	id := ""
	for _, container := range containers {
		for _, name := range container.Names {
			// docker adds a '/' in front of the name
			nameNoLeadingSlash := strings.Replace(name, "/", "", 1)
			if nameNoLeadingSlash == c.Name {
				id = container.ID
			}
		}
	}

	timeout := time.Second * 0
	err = dockerClient.ContainerStop(ctx, id, &timeout)
	if err != nil {
		panic(err)
	}
	removeOptions := types.ContainerRemoveOptions{}
	err = dockerClient.ContainerRemove(ctx, id, removeOptions)
	if err != nil {
		panic(err)
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

func (t Task) executeGithubAction(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {
	panic("Github actions are not implemented yet")
}

func (t Task) executeShellCommand(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {

	fmt.Println("Executing shell command ...")

	// setup command to execute as a slice
	cmd := strings.Split(t.Command, " ")

	// create config for task to execute
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		// Cmd:          []string{"cat", "/etc/os-release"},
		Cmd: cmd,
	}

	execCreateResponse, err := dockerClient.ContainerExecCreate(ctx, containerResponse.ID, execConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println(execCreateResponse)

	execResponse, err := dockerClient.ContainerExecAttach(ctx, execCreateResponse.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}
	defer execResponse.Close()

	// ===

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, execResponse.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			// return execResult, err
			panic(err)
		}
		break

	case <-ctx.Done():
		// return execResult, ctx.Err()
		return 1, nil
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		panic(err)
	}
	// stderr, err := ioutil.ReadAll(&errBuf)
	// if err != nil {
	// panic(err)
	// }
	res, err := dockerClient.ContainerExecInspect(ctx, execCreateResponse.ID)
	if err != nil {
		panic(err)
	}

	exitCode := res.ExitCode
	stdOut := string(stdout)
	// stdErr := string(stderr)

	fmt.Println("Exitcode", exitCode)
	fmt.Println("stdOut", stdOut)
	// fmt.Println("stdErr", stdErr)

	// ====

	return exitCode, nil
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
