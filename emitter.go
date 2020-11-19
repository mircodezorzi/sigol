package main

import (
	"fmt"
	"io/ioutil"
)

func Emit() string {
	file := fmt.Sprintf("service: %s\n\nprovisioner:\n\tname: aws\n\truntime: go1.x\n\nfunctions:", config.Name)

	files, err := ioutil.ReadDir("./cmd")
	check(err)

	for _, f := range files {
		file += fmt.Sprintf("\n\t%s:\n\t\thandler: %s\n\t\tevents:\n\t\t\t- http: any %s", f.Name(), f.Name(), f.Name())
	}

	return file
}
