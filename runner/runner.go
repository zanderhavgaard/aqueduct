package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	imagePullOutput, err := dockerClient.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, imagePullOutput)

	containerName := c.Name
	containerConfig := container.Config{
		Image: c.Image,
		// Cmd:   []string{"cat", "/etc/os-release"},
		Cmd: []string{"tail", "-f", "/dev/null"},
	}

	// create container
	containerResponse, err := dockerClient.ContainerCreate(ctx, &containerConfig, nil, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	// actually start the container
	err = dockerClient.ContainerStart(ctx, containerResponse.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	// Print the ID
	fmt.Println("Running Job", c.Name, "in Container:", containerResponse.ID)

	for index, task := range c.Tasks {
		fmt.Println("Task", index, "...")
		task.execute(ctx, dockerClient, containerResponse)
	}

	// wait for the container to finish running
	// statusCh, errCh := dockerClient.ContainerWait(ctx, containerResponse.ID, container.WaitConditionNotRunning)
	// select {
	// case err := <-errCh:
	// if err != nil {
	// panic(err)
	// }
	// case status := <-statusCh:
	// fmt.Println("Container status:", status)
	// }

	// print logs from the container
	// logsOutput, err := dockerClient.ContainerLogs(ctx, containerResponse.ID, types.ContainerLogsOptions{ShowStdout: true})

	// fmt.Println("foo", logsOutput)
	// stdcopy.StdCopy(os.Stdout, os.Stderr, logsOutput)

	// define the timeout for stopping the containers
	// var timeout time.Duration = time.Second * 30
	var timeout time.Duration = time.Second * 0

	// stop containers, allowing for graceful shutdown
	err = dockerClient.ContainerStop(ctx, containerResponse.ID, &timeout)
	if err != nil {
		panic(err)
	}

	//  remove containers
	err = dockerClient.ContainerRemove(ctx, containerResponse.ID, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}

	return nil
}

func (t Task) execute(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) error {
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
		return nil
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
