package settings

type Settings struct {
	// which platform to emulate
	Platform string
	// which file to run
	Filename string
	// how to execute the pipeline
	ExecutionMode string
	Verbose       bool
	Debug         bool
	SkipImagePull bool
}

// create global settings struct with defaults
var Global Settings = Settings{
	Platform:      "github-actions",
	Filename:      "",
	ExecutionMode: "all",
	Verbose:       false,
	Debug:         false,
	SkipImagePull: false,
}
