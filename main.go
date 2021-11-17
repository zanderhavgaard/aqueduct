package main

import (
	"flag"
	"fmt"

	"github.com/fatih/color"
	"github.com/zanderhavgaard/aqueduct/githubActions"
	"github.com/zanderhavgaard/aqueduct/runner"
	"github.com/zanderhavgaard/aqueduct/settings"
)

// golang does not allow '`' in multiline strings ...
const banner string = "                            _            _   \n" +
	"  __ _  __ _ _   _  ___  __| |_   _  ___| |_ \n" +
	" / _` |/ _` | | | |/ _ \\/ _` | | | |/ __| __|\n" +
	"| (_| | (_| | |_| |  __/ (_| | |_| | (__| |_ \n" +
	" \\__,_|\\__, |\\__,_|\\___|\\__,_|\\__,_|\\___|\\__|\n" +
	"          |_|                                \n"
const subBanner string = "\n~ https://github.com/zanderhavgaard/aqueduct\n\n"

const githubActionsPlatformName string = "github-actions"

func main() {

	// TODO move settings to a settings file
	// execute all steps
	// TODO make a parameter
	// should be able to be specific job and specific step in job
	// settings.Global.ExecutionMode = "all"
	settings.Global.Verbose = true
	settings.Global.Debug = true
	settings.Global.SkipImagePull = true
	settings.Global.RemoveContainers = true
	settings.Global.RemoveConflictingContainers = true
	settings.Global.GracefulContainerShutdown = false

	// print the banner
	// fmt.Println(banner)
	color.Blue(banner)
	color.Cyan(subBanner)

	// get the cli arguments
	parseCliArgs()

	// print settings for this run
	printSettings()

	// the run to execute
	var run runner.Run
	var err error

	// choose which platform to use
	if settings.Global.Platform == githubActionsPlatformName {
		// prepare a run for a githubactions pipeline
		run, err = githubActions.Prepare(settings.Global.Filename)
		if err != nil {
			panic(err)
		}
	} else {
		panic("You must specify the platform with the -platform option.")
	}

	// execute the prepared run
	color.Magenta("--- Run ---")
	fmt.Println("Pipeline:", run.Name)
	fmt.Println("Execution Mode:", settings.Global.ExecutionMode)
	color.Magenta("-----------")
	err = runner.ExecuteRun(run, settings.Global.ExecutionMode)
	if err != nil {
		panic(err)
	}
}

func printSettings() {
	color.Magenta("--- Settings ---")
	fmt.Println("Platform:", settings.Global.Platform)
	fmt.Println("File:", settings.Global.Filename)
	fmt.Println("ExecutionMode:", settings.Global.ExecutionMode)
	color.Magenta("----------------")
}

func parseCliArgs() {
	flag.StringVar(&settings.Global.Platform, "platform", "", "The CI platform to emulate")
	flag.StringVar(&settings.Global.Filename, "file", "", "Path to the .yaml file to read")
	flag.Parse()
}
