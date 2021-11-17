package settings

/*

Platform: which CI pipeline platform to attempt to emulate
- 'github-actions'

Filename: the name of the pipeline file to run

ExecutionMode: how to execute the steps of a given pipeline
- 'all' will execute all steps sequentially

Verbose: enable more verbose prints

Debug: enable debug prints

SkipImagePull: whether to attempt to pull docker images

RemoveConflictingContainers: whether to remove containers with conflicting container names

GracefulContainerShutdown: whether to allow containers to gracefully shutdown or to just kill them when there are no more tasks to do

Remove Containers: whether to remove containers after steps have been executed in them

GitCheckoutMode: whether to clone the repository or bindmount the local repository
- 'clone' will clone the repostory each time
- 'bindmount' will bindmount the local repository to the container

*/

// holds global settings
type Settings struct {
	Platform                    string
	Filename                    string
	ExecutionMode               string
	Verbose                     bool
	Debug                       bool
	SkipImagePull               bool
	RemoveConflictingContainers bool
	GracefulContainerShutdown   bool
	RemoveContainers            bool
	GitCheckoutMode             string
}

// create global settings struct with defaults
var Global Settings = Settings{
	Platform:                    "github-actions",
	Filename:                    "",
	ExecutionMode:               "all",
	Verbose:                     false,
	Debug:                       false,
	SkipImagePull:               false,
	RemoveConflictingContainers: false,
	GracefulContainerShutdown:   true,
	RemoveContainers:            true,
	GitCheckoutMode:             "clone",
}
