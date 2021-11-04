package main

import (
	"flag"
	"fmt"

	"github.com/zanderhavgaard/aqueduct/githubActions"
	"github.com/zanderhavgaard/aqueduct/runner"
)

var platform string
var filename string

const githubActionsPlatformName string = "github-actions"

func main() {
	parseCliArgs()
	// fmt.Println(platform, filename)

	fmt.Println("---")
	fmt.Println("Using platform:", platform)
	fmt.Println("On file:", filename)
	fmt.Println("---")

	if platform == githubActionsPlatformName {
		run, err := githubActions.Prepare(filename)
		if err != nil {
			panic(err)
		}
		executionMode := "all"
		err = runner.ExecuteRun(run, executionMode)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("You must specify the platform with the -platform option.")
	}
}

func parseCliArgs() {
	flag.StringVar(&platform, "platform", "", "The CI platform to emulate")
	flag.StringVar(&filename, "file", "", "Path to the .yaml file to read")
	flag.Parse()
}
