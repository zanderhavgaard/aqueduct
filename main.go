package main

import (
	"flag"
	"fmt"

	"github.com/zanderhavgaard/aqueduct/github"
)

var platform string
var filename string

const githubPlatformName string = "github"

func main() {
	parseCliArgs()
	// fmt.Println(platform, filename)

	fmt.Println("Using platform:", platform, "on file:", filename)
	fmt.Println("---")

	if platform == githubPlatformName {
		_, _ = github.Prepare(filename)
		// fmt.Println(workflow)
	} else {
		fmt.Println("You must specify the platform with the -platform option.")
	}
}

func parseCliArgs() {
	flag.StringVar(&platform, "platform", "", "The CI platform to emulate")
	flag.StringVar(&filename, "file", "", "Path to the .yaml file to read")
	flag.Parse()
}
