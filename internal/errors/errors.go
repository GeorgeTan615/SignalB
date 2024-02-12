package errors

import "fmt"

const (
	RequestDeserializationError = "Error deserializing request"
	DatabaseInsertionError      = "Error inserting data into database"
	DatabaseQueryError          = "Error querying data from database"
)

type ErrorResp struct {
	Message string `json:"message"`
}

func NewErrorResp(msg string) *ErrorResp {
	return &ErrorResp{
		Message: msg,
	}
}

func NewErrorRespWithErr(msg string, err error) *ErrorResp {
	return &ErrorResp{
		Message: fmt.Sprintln(msg, err),
	}
}
