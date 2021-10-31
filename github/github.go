package github

import (
	"fmt"
	"io/ioutil"

	"github.com/goccy/go-yaml"
	"github.com/zanderhavgaard/aqueduct/runner"
)

func Prepare(filename string) (Workflow, error) {
	workflow, err := readYaml(filename)
	if err != nil {
		panic(err)
	}

	run, err := workflow.MakeRun()
	fmt.Println(run)

	return workflow, nil
}

// each step in a job
type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// each job in the workflow
type Job struct {
	Name   string `yaml:"name"`
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

// the entire workflow
type Workflow struct {
	Name string         `yaml:"name"`
	On   string         `yaml:"on"`
	Jobs map[string]Job `yaml:"jobs"`
}

// read a yaml file and parse to a struct
func readYaml(filename string) (Workflow, error) {
	workflow := Workflow{}

	yamlfile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlfile, &workflow)
	if err != nil {
		panic(err)
	}

	return workflow, nil
}

func (w Workflow) MakeRun() (runner.Run, error) {
	containers := []runner.Container{}

	for jobName, jobConfig := range w.Jobs {

		// steps := []runner.Task{}

		// fmt.Println()
		fmt.Println(jobName)
		// fmt.Println(jobConfig.Steps)
		// fmt.Println()

		for _, step := range jobConfig.Steps {
			fmt.Println()
			fmt.Println(step)
			fmt.Println()
		}

		// cont := runner.Container{
		// Name: jobName,
		// }

		// fmt.Println(cont)
		// foo := append(containers, cont)
		// fmt.Println(foo)
	}

	run := runner.Run{
		Name:       w.Name,
		Containers: containers,
	}

	return run, nil
}
