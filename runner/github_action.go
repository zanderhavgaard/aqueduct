package runner

import (
	"context"
	"fmt"
	"regexp"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (t Task) executeGithubAction(ctx context.Context, dockerClient *client.Client, containerResponse container.ContainerCreateCreatedBody) (int, error) {

	// check if uses task is a checkout action
	isCheckoutAction := checkIfCheckoutAction(t.Uses)
	fmt.Println(isCheckoutAction)
	if isCheckoutAction {
		err := checkoutRepo()
		if err != nil {
			panic(err)
		}
	}

	panic("Github actions are not implemented yet")
}

func checkIfCheckoutAction(action string) bool {
	// TODO should be able to handle different versions of the checkout action
	checkoutActionRegex := regexp.MustCompile("actions/checkout@v2")
	found := checkoutActionRegex.MatchString(action)
	if found {
		return true
	} else {
		return false
	}
}
