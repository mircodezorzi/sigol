# sigol: AWS Golang Lambda management tool

![](https://github.com/mircodezorzi/sigol/workflows/Build/badge.svg)

## Installation
```sh
go get github.com/mircodezorzi/sigol
```

## Features
- Automatic lambda creation, management, and deployment
- Automatic REST API Gateways

## Example
```sh
# Create sigol project, pull required dependencies
$ sigol init example

# Create new AWS Lambda `my-function` with dynamo and S3 codegen
$ sigol new my-function --components=dynamodb,s3

# Compile Lambda
$ sigol build my-function

# Upload compiled binary to AWS
$ sigol upload my-function
```

## Other functionalities
- `sigol ls`
- `sigol update`

## Missing features
- better error checking and handling
