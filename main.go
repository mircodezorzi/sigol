package main

import (
	"path"
  "fmt"
  "os"
  "os/exec"
)

type Config struct {
  Path string `yaml:"path"`
}

type Function struct {
  Name     string
	Endpoint string
	Methods  []string
}

var functions []Function
var config Config

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func GoGet(pkg string) error {
  cmd := exec.Command("go", "get", pkg)
  cmd.Env = append(os.Environ(),
    fmt.Sprintf("GOPATH=%s", config.Path),
  )
  return cmd.Start()
}

func GoInit(name string) error {
  cmd := exec.Command("go", "mod", "init", name)
  cmd.Env = append(os.Environ(),
    fmt.Sprintf("GOPATH=%s", config.Path),
  )
  return cmd.Start()
}

func GoBuild(fn string) error {
  cmd := exec.Command("go", "build", ".")
  cmd.Env = append(os.Environ(),
    fmt.Sprintf("GOPATH=%s/src/%s", config.Path, fn),
  )
  return cmd.Start()
}

func GitInit(path string) error {
  cmd := exec.Command("git", "init", path)
  return cmd.Start()
}

func Init(name string) {
	fmt.Printf("initializing project\n")

	check(GitInit(config.Path))
  check(GoGet("github.com/aws/aws-sdk-go"))
  check(GoGet("github.com/aws/aws-lambda-go/lambda"))

	check(GoInit(name))
}

func New(name string) error {
	var err error

  err = os.MkdirAll(fmt.Sprintf("%s/cmd/%s", config.Path, name), 0755)
	check(err)

	file, err := os.Create(fmt.Sprintf("%s/cmd/%s/main.go", config.Path, name))
	defer file.Close()

	return nil
}

func main() {

	config.Path, _ = os.Getwd()

	if len(os.Args) < 2 { return }

	switch c := os.Args[1]; c {
	case "init":
		n, _ := os.Getwd()
		name := path.Base(n)

		if len(os.Args) > 3 {
			config.Path = os.Args[2]
		}

		Init(name)
		break

	case "new":
		if len(os.Args) > 2 {
			New(os.Args[2])
		} else {
			panic("missing argument")
		}
		break

	case "build":
		if len(os.Args) > 2 {
			check(GoBuild(os.Args[2]))
		} else {
			panic("missing argument")
		}
		break
	
	default:
		break
	}
}
