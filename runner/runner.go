package runner

// task to run in a container
type Task struct {
	Name    string
	Command string
}

// a container to run tasks in
type Container struct {
	Name  string
	Tasks []Task
}

// a run of a pipeline
type Run struct {
	Name       string
	Containers []Container
}
