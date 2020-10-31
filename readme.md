# sigol: AWS Golang Lambda management tool

## Example
```sh
# Create sigol project, pull required dependencies
$ sigol init foobar
# Create new AWS Lambda `my-function` with dynamo and S3 codegen
$ sigol new my-function components=dynamodb,s3
# Compile Lambda
$ sigol build my-function
# Upload compiled binary to AWS
$ sigol upload my-function
# Equivalent to build + upload
$ sigol update my-function
```

## Missing features
- automatic creation of HTTP APIs
- component codegen
- graceful error checking
