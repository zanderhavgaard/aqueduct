package runner

import (
	"fmt"
	"os"
	"github.com/go-git/go-git/v5"
)

func checkoutRepo() error {
	repoName := findRepoName()
	fmt.Println(repoName)
	return nil
}

func findRepoName() string {
	// find the name of the git repository

	repoName := ""

	repoPath := "."

	// open existing git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		panic(err)
	}

	// get remotes
	remotes, err := repo.Remotes()

	for _, remote := range remotes {
		fmt.Println("remote", remote)
	}

	os.Exit(0)

	return repoName
}
