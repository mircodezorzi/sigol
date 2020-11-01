package main

import (
	"fmt"
	"os/exec"
)

func GitInit(path string) error {
	cmd := exec.Command("git", "init", path)
	return cmd.Run()
}

func GoInit(path string, name string) error {
	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = path
	return cmd.Run()
}

func GoGet(path string, pkg string) error {
	cmd := exec.Command("go", "get", pkg)
	cmd.Dir = path
	return cmd.Run()
}

func GoBuild(path string, fn string) error {
	cmd := exec.Command("go", "build", "-o", "bin", fmt.Sprintf("./cmd/%s", fn))
	cmd.Dir = path
	return cmd.Run()
}

func GoVendor(path string) error {
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Dir = path
	return cmd.Run()
}
