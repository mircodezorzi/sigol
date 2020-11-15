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
