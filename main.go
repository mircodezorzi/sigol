package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/go-yaml/yaml"

	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Config struct {
	Name    string
	Path    string
	Iam     string `yaml:"iam"`
	Region  string `yaml:"region"`
}

var (
	GOPATH = os.Getenv("GOPATH")

	config Config

	sess *session.Session
	svc *lambda.Lambda
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func awscheck(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				fmt.Println(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				fmt.Println(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				fmt.Println(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeResourceConflictException:
				fmt.Println(lambda.ErrCodeResourceConflictException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				fmt.Println(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				fmt.Println(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
}

// Create zip archive of lambda `fn`
func Zip(fn string) []byte {
	buf := new(bytes.Buffer)

	code, err := ioutil.ReadFile(fmt.Sprintf("%s/bin/%s", config.Path, fn))
	w := zip.NewWriter(buf)

	f, err := w.Create(fn)
	_, err = f.Write([]byte(code))
	check(err)

	w.Close()

	return buf.Bytes()
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

// Returns true if lambda `fn` has already been created, false otherwise
// TODO change to a explicative name
func Exists(fn string) bool {
	input := &lambda.GetFunctionInput{
			FunctionName: aws.String(fn),
	}

	_, err := svc.GetFunction(input)
	return err == nil
}

// Generate all files relative to a lambda
// TODO component codegen here
func New(fn string) {
	var err error

	err = os.MkdirAll(fmt.Sprintf("%s/cmd/%s", config.Path, fn), 0755)
	check(err)

	file, err := os.Create(fmt.Sprintf("%s/cmd/%s/main.go", config.Path, fn))
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf(`package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("%s"))

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body: "ok",
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}`, config.Region))
}

// Upload compiled lambda to AWS
func Upload(fn string) {
	if Exists(fn) {
		input := &lambda.UpdateFunctionCodeInput{
				FunctionName: aws.String(fn),
				ZipFile:      Zip(fn),
		}
		_, err := svc.UpdateFunctionCode(input)
		awscheck(err)
	} else {
		if config.Iam == "" {
			fmt.Errorf("missing IAM role in configuration, cannot create new lambdas")
			return
		}
		input := &lambda.CreateFunctionInput{
			Code:         &lambda.FunctionCode{
				ZipFile:    Zip(fn),
			},
			FunctionName: aws.String(fn),
			Handler:      aws.String("main"),
			Role:         aws.String(config.Iam),
			Runtime:      aws.String("go1.x"),
		}
		result, err := svc.CreateFunction(input)
		awscheck(err)

		arn := *result.FunctionArn
		Gateway(config.Name)
		Resource(fn)
		Method("GET")
		Integration("GET", arn)
	}
}

func List(target string) {
	if target == "--local" {
		fmt.Println("Local Lambdas:")
		files, err := ioutil.ReadDir("./cmd")
		check(err)

		ApiId = CheckForGateway(config.Name)

		for _, f := range files {
			fmt.Printf("%s https://%s.execute-api.%s.amazonaws.com/default/%s\n", f.Name(), ApiId, config.Region, f.Name())
		}
	}
	if target == "--remote" {
		fmt.Println("Remote Lambdas:")
		ApiId = CheckForGateway(config.Name)
		resources := GetPaths()

		for _, r := range resources {
			fmt.Printf("%s https://%s.execute-api.%s.amazonaws.com/default/%s\n", r, ApiId, config.Region, r)
		}
	}
}

func main() {
	config.Path, _ = os.Getwd()
	config.Name = path.Base(config.Path)

	// YAML > JSON
	data, _ := ioutil.ReadFile(config.Path + "/.sigol.yml")
	_ = yaml.Unmarshal([]byte(data), &config)

	if len(os.Args) < 2 { return }
	if config.Iam == "" {
		fmt.Print("missing IAM role in configuration, won't be able to create new lambdas")
	}
	if config.Region == "" {
		panic("missing region in configuration, won't be able to create new lambdas")
	}

	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc = lambda.New(sess, &aws.Config{Region: aws.String(config.Region)})

	switch os.Args[1] {
	case "init":
		/*
		if len(os.Args) > 2 {
			if os.Args[2] == "." {
				config.Path, _ = os.Getwd()
				config.Name = path.Base(config.Path)
			} else {
				config.Name = os.Args[2]
				config.Path = GOPATH + "/src/" + os.Args[2]
			}
		} else {
			panic("missing argument")
		}
		*/
		Init()
		break

	case "new":
		if !IsProject() { return }
		if len(os.Args) > 2 {
			New(os.Args[2])
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
