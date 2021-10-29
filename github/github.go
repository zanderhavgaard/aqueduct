package github

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func Parse(filename string) {
	readYaml(filename)
}

// type actions struct {
	// foo string
// }

func readYaml(filename string) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	data := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFile, &data)

	if err != nil {
		panic(err)
	}

	fmt.Println(data)

	// for k, v := range data {
	// fmt.Println(k, v)
	// }

	// fmt.Println(data["jobs"].(map[string]interfact{}))
}
