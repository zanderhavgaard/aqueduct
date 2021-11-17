package runner

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

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
