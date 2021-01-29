package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	ErrInvalidParams     = errors.New("[err] invalid params")
	ErrInvalidFileFormat = errors.New("[err] invalid file format")
	ErrNotFoundPath      = errors.New("[err] not found path")
	ErrUnknownValue      = errors.New("[err] unknown value")
	ErrNotFoundContext   = errors.New("[err] not found context")
)

// PanicWithRed is to print error message with red color and to exit process.
func PanicWithRed(err error) {
	fmt.Println(color.RedString("%s", err.Error()))
	os.Exit(1)
}
