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
	"github.com/fatih/color"
)

func (t Task) executeShellCommand(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {
	color.Blue(fmt.Sprintf("Executing shell command: %s", t.Command))

	// setup command to execute as a slice
	cmd := strings.Split(t.Command, " ")

	// create config for task to execute
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execCreateResponse, err := dockerClient.ContainerExecCreate(ctx, containerResponse.ID, execConfig)
	if err != nil {
		panic(err)
	}
	execID := execCreateResponse.ID

	execResponse, err := dockerClient.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}
	defer execResponse.Close()

	// ===
	// TODO figure out how to stream the output while the command runs instead of dumping the whole output after

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

	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		panic(err)
	}

	res, err := dockerClient.ContainerExecInspect(ctx, execID)
	if err != nil {
		panic(err)
	}

	exitCode := res.ExitCode
	stdOutString := string(stdout)
	stdErrString := string(stderr)

	if exitCode == 0 {
		color.Green("Command executed successfully with exitcode 0")
		color.Green("stdout from the command:")
		color.Magenta("---")
		fmt.Println(stdOutString)
		color.Magenta("---")
		if stdErrString != "" {
			fmt.Println("---")
			fmt.Println("stderr from the command:")
			fmt.Println(stdErrString)
		}
	} else {
		color.Red(fmt.Sprintf("Command exitted with a non-zero exitcode: %d", exitCode))
		color.Red("stdout from the command:")
		color.Magenta("---")
		fmt.Println(stdOutString)
		color.Magenta("---")
		if stdErrString != "" {
			color.Red("---")
			color.Red("stderr from the command:")
			fmt.Println(stdErrString)
		}
	}

	// ====

	return exitCode, nil
}
