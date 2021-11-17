package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/zanderhavgaard/aqueduct/settings"
)

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
