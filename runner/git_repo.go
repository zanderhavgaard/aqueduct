package runner

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
)

func checkoutRepo() error {
	repoName := findRepoName()
	fmt.Println(repoName)
	return nil
}

func findRepoName() string {
	// TODO there should be a simpler way to extract the remote name
	// should be possible using the git module and not using regex,
	// I just can't figure out how to get remote name from the struct...

	// find the name of the git repository
	repoName := ""
	repoPath := "."

	// open existing git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		panic(err)
	}

	// get the origin remote
	origin, err := repo.Remote("origin")
	panicIfErr(err)

	// convert to string
	originString := origin.String()

	// get the <username>/<reponame> of the github repository
	// this should work regardless of the repo was cloned with ssh or http
	re := regexp.MustCompile("[github.com/|git@github.com:](\\w+/\\w+)\\ \\(fetch\\)")
	match := re.FindStringSubmatch(originString)

	// first element in slice is the entire match of the regex
	// second element is the first match of the regex group
	// therefore we only want what is in the group
	repoName = match[1]

	return repoName
}
