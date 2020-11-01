package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

var ApiId string
var RootId string
var ResourceId string
var MethodId string

func Gateway(name string) {
	ApiId = CheckForGateway(name)

	if ApiId == "" {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.CreateRestApiInput{
			Name: aws.String(name),
		}
		result, err := svc.CreateRestApi(input)
		awscheck(err)

		ApiId = *result.Id
	}
}

func CheckForGateway(name string) string {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetRestApisInput{ }
	result, err := svc.GetRestApis(input)
	awscheck(err)

	for _, j := range result.Items {
		if *j.Name == name {
			return *j.Id
		}
	}

	return ""
}

func Resource(name string) {
	RootId = CheckForResource("/")
	ResourceId = CheckForResource(name)

	if ResourceId == "" {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.CreateResourceInput{
			PathPart:  aws.String(name),
			ParentId:  aws.String(RootId),
			RestApiId: aws.String(ApiId),
		}
		result, err := svc.CreateResource(input)
		awscheck(err)

		ResourceId = *result.Id
	}
}

func CheckForResource(name string) string {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(ApiId),
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

func CheckForMethod(method string) bool {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.GetMethodInput{
		HttpMethod: aws.String(method),
		ResourceId: aws.String(ResourceId),
		RestApiId:  aws.String(ApiId),
	}
	_, err := svc.GetMethod(input)
	return err == nil
}

func Method(method string) {
	if !CheckForMethod(method) {
		svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
		input := &apigateway.PutMethodInput{
			AuthorizationType: aws.String("NONE"),
			HttpMethod:        aws.String(method),
			ResourceId:        aws.String(ResourceId),
			RestApiId:         aws.String(ApiId),
		}
		_, err := svc.PutMethod(input)
		awscheck(err)
	}
}

func Integration(method string, arn string) {
	svc := apigateway.New(sess, &aws.Config{Region: aws.String(config.Region)})
	input := &apigateway.PutIntegrationInput{
		HttpMethod:            aws.String(method),
		IntegrationHttpMethod: aws.String(method),
		ResourceId:            aws.String(ResourceId),
		RestApiId:             aws.String(ApiId),
		Type:                  aws.String("AWS"),
		Uri:                   aws.String(fmt.Sprintf("arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations", config.Region, arn)),
	}
	_, err := svc.PutIntegration(input)
	awscheck(err)
}
