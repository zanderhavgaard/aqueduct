package githubActions

import (
	"io/ioutil"

	"github.com/goccy/go-yaml"
	"github.com/zanderhavgaard/aqueduct/runner"
)

func Prepare(filename string) (runner.Run, error) {
	workflow, err := readYaml(filename)
	if err != nil {
		panic(err)
	}

	run, err := workflow.MakeRun()

	return run, nil
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

		tasks := []runner.Task{}

		for _, step := range jobConfig.Steps {
			task := runner.Task{
				Name:    step.Name,
				Command: step.Run,
			}
			tasks = append(tasks, task)
		}

		// default image
		image := "ubuntu:latest"

		if jobConfig.RunsOn == "ubuntu-latest" {
			image = "ubuntu:latest"
		}

		container := runner.Container{
			Name:  jobName,
			Image: image,
			Tasks: tasks,
		}

		containers = append(containers, container)
	}

	run := runner.Run{
		Name:       w.Name,
		Containers: containers,
	}

	return run, nil
}
