package runner

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (t Task) executeGithubAction(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {
	panic("Github actions are not implemented yet")
}
