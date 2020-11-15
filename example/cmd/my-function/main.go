package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var (
	sess *session.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	REGION = "eu-south-1"

	dynamodb = dynamodb.New(sess, &aws.Config{Region: aws.String(REGION)}
	ecs = ecs.New(sess, &aws.Config{Region: aws.String(REGION)}
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body: "ok",
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}