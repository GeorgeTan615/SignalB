package errors

const (
	RequestDeserializationError = "error deserializing request"
	DatabaseInsertionError      = "error inserting data into database"
	DatabaseQueryError          = "error querying data from database"
)

type ErrorResp struct {
	Message string `json:"message"`
}

func NewErrorResp(err error) *ErrorResp {
	return &ErrorResp{
		Message: err.Error(),
	}
}
