package main

import (
	"github.com/zanderhavgaard/aqueduct/github"
)

func main() {
	pipelineFile := "examples/hello-world.yaml"
	github.Parse(pipelineFile)
}
