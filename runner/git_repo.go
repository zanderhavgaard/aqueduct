package runner

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/zanderhavgaard/aqueduct/settings"
)

func checkoutRepo() error {
	// either attempt to clone the repository or bindmount the local repository
	if settings.Global.GitCheckoutMode == "clone" {
		repoName := findRepoName()
		cloneRepo(repoName)
	} else if settings.Global.GitCheckoutMode == "bindmount" {
		fmt.Println("bindmount")
		panic("not implemented")
	} else {
		panic("unknown git checkout mode, must be one of 'clone', 'bindmount'")
	}

	return nil
}

func cloneRepo(repoName string) {
	// TODO figure how to handle auth for private repos

	platformURL := ""
	// TODO add other relevant platforms
	if settings.Global.Platform == "github-actions" {
		platformURL = "https://github.com/"
	}
	repoURL := platformURL + repoName

	// find user home dir
	userHome, err := os.UserHomeDir()
	panicIfErr(err)

	// build path to repo cache
	clonePath := userHome + "/.cache/aqueduct/repos/" + repoName

	// TODO be smarter about using cached repo instead of just deleting and recloning

	// check if directory already exists
	_, err = os.Stat(clonePath)
	if os.IsNotExist(err) {
		fmt.Println("Creating directories for repository")
		// create dirs
		os.MkdirAll(clonePath, 0755)

	} else {
		fmt.Println("Removing existing cached repo")
		// remove directory and files
		err = os.RemoveAll(clonePath)
		panicIfErr(err)
	}

	fmt.Println("Cloning repository:", repoURL, "to directory:", clonePath)

	// setup the options for the clone
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}
	// clone the repository form remote
	_, err = git.PlainClone(clonePath, false, cloneOptions)
	panicIfErr(err)
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
