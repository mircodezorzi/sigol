package main

import (
	"fmt"
)

type CannotZipFileError struct {
	filepath string
}

func (e *CannotZipFileError) Error() string {
	return fmt.Sprintf("error while zipping %s", e.filepath)
}

type LambdaExistsError struct {
	method string
	name   string
}

func (e *LambdaExistsError) Error() string {
	return fmt.Sprintf("lambda %s (%s) already exists", e.name, e.method)
}
