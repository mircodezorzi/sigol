package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type Api struct {
	ApiId      string
	ResourceId string

	// Gateway Name
	Gateway    string
}

func NewApi(name string) Api {
	api := Api{
		Gateway: name,
	}
	api.ApiId = api.CheckForGateway(name)
	return api
}

// \return Whether a Lambda `fn` already exists.
// TODO iterate throguht paginator when necessary
func LambdaExists(fn string) bool {
	svc := lambda.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(fn),
	}

	_, err := svc.GetFunction(input)
	return err == nil
}

// \brief Ensure gateway exists by creating one if it doesn't already exist.
func (api *Api) NewLambda(name string, method string) error {
	if LambdaExists(name) {
		return &LambdaExistsError{method, name}
	}

	compressed, err := Zip(name)
	if err != nil {
		return err
	}

	svc := lambda.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String(name),
		Handler:      aws.String("main"),
		Role:         aws.String(config.Iam),
		Runtime:      aws.String("go1.x"),
		Code:         &lambda.FunctionCode{
			ZipFile:    compressed,
		},
	}

	result, err := svc.CreateFunction(input)
	if err != nil {
		return err
	}

	arn := *result.FunctionArn

	api.EnsureGateway(api.Gateway)
	api.EnsureResource(name)
	api.EnsureMethod(method)
	api.EnsureIntegration(method, arn)

	return nil
}

func (api *Api) UpdateLambda(name string) error {
	compressed, err := Zip(name)
	if err != nil {
		return err
	}

	svc := lambda.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(name),
			ZipFile:      compressed,
	}
	_, err = svc.UpdateFunctionCode(input)
	return err
}

// \brief Ensure gateway exists by creating one if it doesn't already exist.
// \return Gateway ID.
func (api *Api) EnsureGateway(name string) {
	if api.ApiId == "" {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.CreateRestApiInput{
			Name: aws.String(name),
		}
		result, err := svc.CreateRestApi(input)
		awscheck(err)

		api.ApiId = *result.Id
	}
}

// \return Gateway ID.
// TODO iterate throguht paginator when necessary
func (api *Api) CheckForGateway(name string) string {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetRestApisInput{}
	result, err := svc.GetRestApis(input)
	awscheck(err)

	for _, j := range result.Items {
		if *j.Name == name {
			return *j.Id
		}
	}

	return ""
}

// \brief Ensure resource exists by creating one if it doesn't already exist.
// \return Resource ID.
func (api *Api) EnsureResource(name string) {
	rootId := api.CheckForResource("/")
	api.ResourceId = api.CheckForResource(name)

	if api.ResourceId == "" {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.CreateResourceInput{
			PathPart:  aws.String(name),
			ParentId:  aws.String(rootId),
			RestApiId: aws.String(api.ApiId),
		}
		result, err := svc.CreateResource(input)
		awscheck(err)

		api.ResourceId = *result.Id
	}
}

// \return Resource ID
// TODO iterate throguht paginator when necessary
func (api *Api) CheckForResource(name string) string {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(api.ApiId),
	}
	result, err := svc.GetResources(input)
	awscheck(err)

	for _, j := range result.Items {
		if j.PathPart != nil {
			if *j.PathPart == name {
				return *j.Id
			}
		} else if j.Path != nil {
			if *j.Path == name {
				return *j.Id
			}
		}
	}

	return ""
}

// \brief Ensure method exists by creating one if it doesn't already exist.
// \return Method ID.
func (api *Api) EnsureMethod(method string) {
	if !api.CheckForMethod(method) {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.PutMethodInput{
			AuthorizationType: aws.String("NONE"),
			HttpMethod:        aws.String(method),
			ResourceId:        aws.String(api.ResourceId),
			RestApiId:         aws.String(api.ApiId),
		}
		_, err := svc.PutMethod(input)
		awscheck(err)
	}
}

// \return Method ID
// TODO iterate throguht paginator when necessary
func (api *Api) CheckForMethod(method string) bool {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetMethodInput{
		HttpMethod: aws.String(method),
		ResourceId: aws.String(api.ResourceId),
		RestApiId:  aws.String(api.ApiId),
	}
	_, err := svc.GetMethod(input)
	return err == nil
}

// \brief Ensure integration exists by creating one if it doesn't already exist.
// \return Method ID.
func (api *Api) EnsureIntegration(method string, arn string) {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.PutIntegrationInput{
		HttpMethod:            aws.String(method),
		IntegrationHttpMethod: aws.String(method),
		ResourceId:            aws.String(api.ResourceId),
		RestApiId:             aws.String(api.ApiId),
		Type:                  aws.String("AWS"),
		Uri:                   aws.String(fmt.Sprintf("arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations", config.Region, arn)),
	}
	_, err := svc.PutIntegration(input)
	awscheck(err)
}

// \return All paths in an api.
func (api *Api) GetPaths() []string {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(api.ApiId),
	}
	result, err := svc.GetResources(input)
	awscheck(err)

	var results []string

	for _, i := range result.Items {
		if i.PathPart != nil {
			results = append(results, *i.PathPart)
		}
	}

	return results
}
