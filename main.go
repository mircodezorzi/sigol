package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/awserr"


	"github.com/go-yaml/yaml"
)

var template = `package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"%s
)%s

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body: "ok",
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}`

type Config struct {
	// Project Names
	Name    string
	// Project Path
	Path    string

	Iam     string `yaml:"iam"`
	Region  string `yaml:"region"`
}

var (
	GOPATH = os.Getenv("GOPATH")

	config Config

	sess *session.Session
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func awscheck(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return
	}
}

// Create in-memory archive of Lambda `fn`
func Zip(fn string) ([]byte, error) {
	buf := new(bytes.Buffer)

	code, err := ioutil.ReadFile(fmt.Sprintf("%s/bin/%s", config.Path, fn))
	w := zip.NewWriter(buf)

	f, err := w.Create(fn)
	_, err = f.Write([]byte(code))
	if err != nil {
		return []byte(nil), &CannotZipFileError{fn}
	}

	w.Close()

	return buf.Bytes(), nil
}

// Return true if current directory is a sigol path, false otherwise
func IsProject() bool {
	if _, err := os.Stat(config.Path + "/.sigol.yml"); os.IsNotExist(err) {
		return false;
	}
	return true;
}

// Initialize a sigol project
//
// Created directories:
// - bin: compiled binaries
// - cmd: source code
func Init() {
	fmt.Printf("initializing project\n")

	check(os.MkdirAll(config.Path, 0755))
	check(os.MkdirAll(config.Path + "/bin", 0755))
	check(os.MkdirAll(config.Path + "/cmd", 0755))

	check(GitInit(config.Path))
	check(GoInit(config.Path, config.Name))
	check(GoGet(config.Path, "github.com/aws/aws-lambda-go/events"))
	check(GoGet(config.Path, "github.com/aws/aws-lambda-go/lambda"))
	check(GoGet(config.Path, "github.com/aws/aws-sdk-go/aws"))
	check(GoGet(config.Path, "github.com/aws/aws-sdk-go/aws/session"))
	check(GoGet(config.Path, "github.com/aws/aws-sdk-go/service/dynamodb"))


	file, err := os.Create(fmt.Sprintf("%s/.sigol.yml", config.Path))
	defer file.Close()
	check(err)
}

func FormatTemplate(services []string) string {
	var (
		vars    string
		imports string
	)

	for _, j := range services {
		imports += fmt.Sprintf("\n\t\"github.com/aws/aws-sdk-go/service/%s\"", j)
	}

	for _, j := range services {
		vars += fmt.Sprintf("\n\t%s = %s.New(sess, &aws.Config{Region: aws.String(REGION)}", j, j)
	}

	if imports != "" {
		imports = "\n\n\t\"github.com/aws/aws-sdk-go/aws\"\n\t\"github.com/aws/aws-sdk-go/aws/session\"\n" + imports
	}

	if vars != "" {
		vars = fmt.Sprintf(`

var (
	sess *session.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	REGION = "%s"
%s
)`, config.Region, vars)
	}

	return fmt.Sprintf(template, imports, vars)
}

// Generate all files relative to a lambda
// TODO component codegen here
func New(fn string, components []string) {
	var err error

	err = os.MkdirAll(fmt.Sprintf("%s/cmd/%s", config.Path, fn), 0755)
	check(err)

	file, err := os.Create(fmt.Sprintf("%s/cmd/%s/main.go", config.Path, fn))
	defer file.Close()

	_, err = file.WriteString(FormatTemplate(components))
}

// Upload compiled lambda to AWS
func Upload(fn string) {
	api := NewApi(config.Name)

	if LambdaExists(fn) {
		err := api.UpdateLambda(fn)
		awscheck(err)
	} else {
		if config.Iam == "" {
			fmt.Errorf("missing IAM role in configuration, cannot create new lambdas")
			return
		}
		err := api.NewLambda(fn, "GET")
		awscheck(err)
	}
}

func List(target string) {
	api := NewApi(config.Name)

	if target == "--local" {
		fmt.Println("Local Lambdas:")
		files, err := ioutil.ReadDir("./cmd")
		check(err)

		for _, f := range files {
			fmt.Printf("%s https://%s.execute-api.%s.amazonaws.com/default/%s\n", f.Name(), api.ApiId, config.Region, f.Name())
		}
	}
	if target == "--remote" {
		fmt.Println("Remote Lambdas:")
		resources := api.GetPaths()

		for _, r := range resources {
			fmt.Printf("%s https://%s.execute-api.%s.amazonaws.com/default/%s\n", r, api.ApiId, config.Region, r)
		}
	}
}

func Help() {
	fmt.Print(`Usage: sigol [commands...]

Commands:
	help	Print this message
	init	Create new sigol project
	new	Create new Lambda
	upload	Upload Lambda to AWS
	build	Build golang Lambda
	update	Equivalent to build and update
	ls	List --local or --remote Lambdas
	gen	Generate serverless.yml

Examples:
	sigol init example
	sigol new my-function --components=dynamodb,s3
	sigol build my-function
	sigol upload my-function
`)
	os.Exit(1)
}

func main() {

	config.Path, _ = os.Getwd()
	config.Name = path.Base(config.Path)

	// YAML > JSON
	data, _ := ioutil.ReadFile(config.Path + "/.sigol.yml")
	_ = yaml.Unmarshal([]byte(data), &config)

	if len(os.Args) < 2 {
		Help()
	}

	if config.Region == "" {
		fmt.Println("missing region in configuration, quitting")
		os.Exit(1)
	}

	if config.Iam == "" {
		fmt.Println("missing IAM role in configuration, won't be able to create new lambdas")
	}

	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	switch os.Args[1] {
	case "help":
		Help()
		break

	case "init":
		if IsProject() {
			fmt.Printf("Already in a sigol project.")
			return
		}
		Init()
		break

	case "new":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			var components []string
			if len(os.Args) > 3 {
				split := strings.Split(os.Args[3], "=")
				if len(split) > 1 && len(split) < 3 {
					if split[0] == "--components" || split[0] == "-c" {
						components = strings.Split(split[1], ",")
					}
				}
			}
			New(os.Args[2], components)
		} else {
			panic("missing argument")
		}
		break

	case "ls":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			if os.Args[2] == "--local" || os.Args[2] == "--remote" {
				List(os.Args[2])
			} else {
				panic("invalid argument")
			}
		} else {
			List("--local")
		}
		break

	case "build":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			check(GoVendor(config.Path))
			check(GoBuild(config.Path, os.Args[2]))
		} else {
			panic("missing argument")
		}
		break

	case "gen":
		if !IsProject() { return }
		fmt.Printf(Emit())
		break

	case "upload":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			Upload(os.Args[2])
		} else {
			panic("missing argument")
		}
		break

	case "update":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			check(GoBuild(config.Path, os.Args[2]))
			Upload(os.Args[2])
		} else {
			panic("missing argument")
		}
		break

	default:
		break
	}
}
