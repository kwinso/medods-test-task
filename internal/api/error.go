package api

var (
	InternalServerErrorResponse = ErrorResponse{Error: "Internal Server Error"}
	BadRequestResponse          = ErrorResponse{Error: "Bad Request"}
	UnauthorizedResponse        = ErrorResponse{Error: "Unauthorized"}
)

// ErrorResponse holds a generic error response
// @Description	Generic error response
type ErrorResponse struct {
	// Contains the error message
	Error string `json:"error"`
}
