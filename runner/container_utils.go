package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/zanderhavgaard/aqueduct/settings"
)

func (c Container) copyFilesFromHost(hostPath string, containerPath string) {
	// copy files from hostPath to containerPath

	panic("not implemented")

	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	panicIfErr(err)

	// TODO finish implementing
	content := nil
	options := types.CopyToContainerOptions{}
	err = dockerClient.CopyToContainer(ctx, c.ID, containerPath, content, options)

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
	options := types.ContainerListOptions{All: true}
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
	options := types.ContainerListOptions{All: true}
	containers, err := dockerClient.ContainerList(ctx, options)
	panicIfErr(err)

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
	if id == "" {
		panic("Error: Could not find id of container")
	}

	// for manually stopping the container before removing it
	// timeout := time.Second * 0
	// err = dockerClient.ContainerStop(ctx, id, &timeout)
	// panicIfErr(err)

	// instead of manually stopping container first we can just force remove,
	// saves having to check if the container is running first
	removeOptions := types.ContainerRemoveOptions{Force: true}
	err = dockerClient.ContainerRemove(ctx, id, removeOptions)
	panicIfErr(err)

	return nil
}
