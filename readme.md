# sigol: AWS Golang Lambda management tool

## Example
```sh
# Create sigol project, pull required dependencies
$ sigol init foobar
# Create new AWS Lambda `token` with dynamo and S3 codegen
$ sigol new token components=dynamodb,s3
# Compile Lambda
$ sigol build token
# Generate zip file and upload to AWS
$ sigol build token
# Equivalent to build + upload
$ sigol update token
```

## Missing features
- automatic creation of HTTP APIs
- component codegen
- graceful error checking
